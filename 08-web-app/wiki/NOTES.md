# 08-wiki — Tutorial Notes

Source: https://go.dev/doc/articles/wiki/

---

## What this tutorial covers

This is the most complete single-file tutorial in the series. It builds a
working multi-page wiki from scratch, introducing several concepts in sequence:

| Concept | Where |
|---|---|
| Struct methods for persistence | `Page.save()`, `loadPage()` |
| `net/http` handler pattern | `viewHandler`, `editHandler`, `saveHandler` |
| `html/template` — safe HTML rendering | `view.html`, `edit.html`, `renderTemplate()` |
| Template caching with `template.Must` | `var templates = ...` |
| `http.Redirect`, `http.Error`, `http.NotFound` | throughout handlers |
| `regexp` for input validation | `var validPath = ...` |
| Closures as middleware | `makeHandler()` |

---

## Project structure

```
08-wiki/
├── go.mod
├── wiki.go       ← full application
├── view.html     ← template for viewing a page
├── edit.html     ← template for editing a page
├── NOTES.md      ← this file
└── *.txt         ← page files created at runtime (e.g. FrontPage.txt)
```

The `.txt` files are created in the same directory as the binary when pages
are saved. They are the "database" for this wiki.

---

## Running the server

```bash
go run wiki.go
```

Then open your browser at:

| URL | What you see |
|---|---|
| `http://localhost:8080/view/FrontPage` | Redirects to edit (page doesn't exist yet) |
| `http://localhost:8080/edit/FrontPage` | Edit form for FrontPage |
| `http://localhost:8080/view/FrontPage` | View the page after saving |

Stop the server with `Ctrl+C`. Page files persist on disk between runs.

---

## Key concepts explained

### Methods on structs — persistence without a database

```go
func (p *Page) save() error {
    return os.WriteFile(p.Title+".txt", p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
    body, err := os.ReadFile(title + ".txt")
    ...
}
```

The wiki stores each page as a `.txt` file. `save` and `loadPage` are the
entire persistence layer. No database, no ORM.

`(p *Page)` is a **pointer receiver** — the method operates on the original
struct, not a copy. Required here because `save` reads `p.Body` (read-only
would also work with a value receiver, but pointer receivers are conventional
when the struct might be mutated in the future).

---

### `net/http` handler pattern

A handler is any function with the signature:

```go
func(w http.ResponseWriter, r *http.Request)
```

- `w http.ResponseWriter` — write your response here (headers + body)
- `r *http.Request` — read the request from here (URL, method, body, headers)

Register handlers with `http.HandleFunc(pattern, handler)`. The pattern is a
path prefix — `"/view/"` matches `/view/anything`.

`http.ListenAndServe(addr, nil)` starts the server. `nil` means "use the
default mux" (the one populated by `http.HandleFunc`). It blocks forever and
only returns on error, so `log.Fatal` is the correct wrapper.

---

### `html/template` — why not `fmt.Fprintf`?

The early version of `editHandler` used raw `fmt.Fprintf` to write HTML:

```go
// UNSAFE — hard-coded HTML with user data interpolated directly
fmt.Fprintf(w, "<h1>Editing %s</h1><textarea>%s</textarea>", title, body)
```

Problems:
1. Hard to read and maintain
2. **XSS risk** — if `title` or `body` contains `<script>`, it gets sent raw

`html/template` solves both:
- HTML lives in separate `.html` files
- Any value interpolated with `{{.Field}}` is **automatically HTML-escaped**:
  `<` → `&lt;`, `>` → `&gt;`, `"` → `&#34;`, etc.

```go
// SAFE — html/template escapes .Body automatically
{{printf "%s" .Body}}
```

---

### Template caching with `template.Must`

```go
// Parsed ONCE at startup — not on every request
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
```

Without caching, `renderTemplate` would call `ParseFiles` on every request —
reading and parsing files thousands of times per second under load.

`template.Must` is a convenience wrapper:
- If `ParseFiles` returns an error → **panic immediately** at startup
- If `ParseFiles` succeeds → return the `*Template` unchanged

Panicking at startup is correct here: a server with broken templates cannot
serve anything useful, and it's better to crash loudly than to silently
serve errors on every request.

To render a specific template by name:

```go
templates.ExecuteTemplate(w, "view.html", p)
// "view.html" is the base filename registered by ParseFiles
```

---

### `http.Redirect`, `http.Error`, `http.NotFound`

| Function | HTTP status | When to use |
|---|---|---|
| `http.Redirect(w, r, url, code)` | 302 / 301 | Send browser to a different URL |
| `http.Error(w, msg, code)` | any (usually 500) | Report an error to the client |
| `http.NotFound(w, r)` | 404 | Resource doesn't exist |

`http.StatusFound` (302) is used for POST → redirect → GET, the standard
web form submission pattern. After saving a page, redirect to `/view/` so
the browser doesn't re-POST on refresh.

---

### Input validation with `regexp`

```go
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
```

Without validation, a user could request `/view/../../etc/passwd` and the
server would try to read that file. The regex anchors (`^` and `$`) and the
character class `[a-zA-Z0-9]+` ensure only safe alphanumeric titles reach
the filesystem.

`regexp.MustCompile` panics on an invalid pattern — fine for package-level
variables because a bad pattern is a programmer error caught at startup.

`FindStringSubmatch` returns a slice of strings:

```
m[0] = "/view/SomePage"   ← full match
m[1] = "view"             ← first capture group  (action)
m[2] = "SomePage"         ← second capture group (title)
```

---

### Closures as middleware — `makeHandler`

This is the most conceptually dense part of the tutorial.

**The problem:** every handler needs to validate the URL and extract the title.
Repeating that logic in three handlers is duplication.

**The solution:** a higher-order function that wraps any handler:

```go
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])   // call the real handler with the validated title
    }
}
```

The inner `func(w, r)` is a **function literal** (anonymous function).
It is a **closure** because it captures `fn` from the outer scope — even
after `makeHandler` has returned, the closure remembers `fn`.

Usage:

```go
http.HandleFunc("/view/", makeHandler(viewHandler))
```

`makeHandler(viewHandler)` returns a new `http.HandlerFunc`. That function
is what `net/http` calls on each request. When it runs, it validates the URL
then calls `viewHandler(w, r, title)` — the real logic.

This is the standard Go pattern for middleware: a function that takes a
handler and returns a new handler with additional behaviour wrapping it.

---

### Evolution of the code through the tutorial

The tutorial deliberately shows bad practices first, then improves them:

| Step | Problem introduced | Fix applied |
|---|---|---|
| `loadPage` v1 | ignores error with `_` | returns `(*Page, error)` |
| `editHandler` v1 | hard-coded HTML string | use `html/template` |
| `renderTemplate` v1 | ignores template errors | check both `ParseFiles` and `Execute` errors |
| `renderTemplate` v2 | calls `ParseFiles` every request | cache with `template.Must` at startup |
| handlers v1 | each duplicates path/title extraction | centralise in `makeHandler` closure |

Reading the commit history of the tutorial (or the intermediate code listings
linked in the article) shows this refactoring process in action.
