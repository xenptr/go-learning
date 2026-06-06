// Package main implements a simple in-filesystem wiki web application.
//
// Source: https://go.dev/doc/articles/wiki/
//
// DONE: Store templates in tmpl/ and page data in data/
// DONE: Add a handler to make the web root redirect to /view/FrontPage
// DONE: Spruce up the page templates by making them valid HTML and adding some CSS rules
// DONE: Implement inter-page linking by converting instances of [PageName]
//       to <a href="/view/PageName">PageName</a> using regexp.ReplaceAllFunc
//
// Run:
//
//	go run wiki.go
//
// Then open http://localhost:8080/ — it will redirect to /view/FrontPage.
// Visit /view/AnyName to create a new page (redirects to the edit form).
//
// Pages are stored as data/<Title>.txt
// Templates are loaded from tmpl/edit.html and tmpl/view.html
package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// ============================================================
// DONE: Store page data in data/
// ============================================================
//
// dataDir is the directory where page .txt files are stored.
// Using a subdirectory keeps page files separate from source code and
// makes the project layout cleaner.
const dataDir = "data"

// ============================================================
// Data model
// ============================================================

// Page represents a single wiki page stored on disk.
type Page struct {
	Title string
	Body  []byte

	// DONE: Inter-page linking — BodyHTML holds the body with [PageName]
	// patterns already converted to HTML links. It is populated by
	// loadPage and used by the view template instead of raw Body.
	// Declared as template.HTML so html/template does NOT escape it —
	// we trust our own regexp conversion, not user-supplied HTML.
	BodyHTML template.HTML
}

// save writes the page body to data/<Title>.txt.
//
// DONE: Store page data in data/
// os.MkdirAll creates the data/ directory if it doesn't exist yet,
// making the first save on a fresh checkout work without manual setup.
func (p *Page) save() error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}
	filename := filepath.Join(dataDir, p.Title+".txt")
	return os.WriteFile(filename, p.Body, 0600)
}

// linkPattern matches wiki cross-links of the form [PageName].
// Capture group 1 captures the page name (alphanumeric only, matching
// the same character class as validPath so only linkable pages are linked).
//
// DONE: Inter-page linking
var linkPattern = regexp.MustCompile(`\[([a-zA-Z0-9]+)\]`)

// loadPage reads a page from data/<title>.txt and returns a *Page.
// It also pre-processes the body to convert [PageName] patterns into
// HTML anchor tags, storing the result in Page.BodyHTML.
//
// DONE: Store page data in data/
// DONE: Inter-page linking
func loadPage(title string) (*Page, error) {
	filename := filepath.Join(dataDir, title+".txt")
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// DONE: Inter-page linking
	// regexp.ReplaceAllFunc iterates over every match of linkPattern and
	// calls the provided function with the matched bytes, replacing the
	// match with the function's return value.
	//
	// match is e.g. []byte("[GoLang]")
	// linkPattern.FindSubmatch extracts the inner name "GoLang"
	// We build an <a> tag and return it as bytes.
	linkedBody := linkPattern.ReplaceAllFunc(body, func(match []byte) []byte {
		// Extract the page name from inside the brackets.
		// FindSubmatch returns [fullMatch, group1], so [1] is the name.
		sub := linkPattern.FindSubmatch(match)
		if sub == nil {
			return match // shouldn't happen, but be safe
		}
		name := string(sub[1])
		// Build the anchor tag. This HTML is controlled entirely by our
		// code (name is validated alphanumeric), not user-supplied, so
		// it is safe to wrap in template.HTML below.
		return []byte(`<a href="/view/` + name + `">` + name + `</a>`)
	})

	return &Page{
		Title:    title,
		Body:     body,
		BodyHTML: template.HTML(linkedBody), // mark as safe, pre-escaped HTML
	}, nil
}

// ============================================================
// Template caching
// ============================================================

// templates is parsed once at startup with template.Must.
//
// template.Must panics if ParseFiles returns an error. This is intentional:
// if the HTML templates are missing or malformed, the server cannot function
// at all, so a panic at startup is more informative than a runtime error on
// the first request.
//
// ParseFiles names each template after its base filename (e.g. "edit.html",
// "view.html"). ExecuteTemplate refers to them by these names.
var templates = template.Must(
	template.ParseFiles("tmpl/edit.html", "tmpl/view.html"),
)

// renderTemplate executes the named template (e.g. "view") against the Page p
// and writes the result to w. The ".html" extension is appended to match the
// names registered by ParseFiles.
//
// http.Error sends an HTTP 500 response if either step fails, preventing the
// server from writing a partial response and hanging the client.
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ============================================================
// Input validation
// ============================================================

// validPath is a compiled regexp that matches URLs of the form:
//   /(edit|save|view)/PageTitle
// where PageTitle contains only alphanumeric characters.
//
// regexp.MustCompile panics on an invalid pattern (fine for a package-level
// variable — a bad pattern is a programmer error, not a runtime condition).
//
// Using this regex prevents directory traversal and other path-based attacks:
// a title like "../../etc/passwd" would fail the match and get a 404.
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// ============================================================
// Middleware: makeHandler (closure pattern)
// ============================================================

// makeHandler is a higher-order function that wraps a handler expecting a
// validated title string into a standard http.HandlerFunc.
//
// This is the "closure as middleware" pattern:
//
//  1. makeHandler receives fn — one of viewHandler, editHandler, saveHandler —
//     as a function value.
//
//  2. It returns a new function (a closure) that:
//     a. Validates the URL path against validPath.
//     b. Extracts the page title (second capture group, m[2]).
//     c. Calls fn with the validated title.
//
//  3. The returned closure "closes over" fn — it captures the variable from
//     the enclosing scope and carries it with it even after makeHandler returns.
//
// Without this pattern, every handler would need its own copy of the
// validation boilerplate. The closure centralises it in one place.
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// FindStringSubmatch returns nil if the path doesn't match.
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		// m[0] = full match, m[1] = action (view/edit/save), m[2] = page title
		fn(w, r, m[2])
	}
}

// ============================================================
// Handlers
// ============================================================

// Each handler now receives a pre-validated title string instead of
// extracting and validating it itself. Separation of concerns: validation
// lives in makeHandler, business logic lives in the handler.

// viewHandler serves GET /view/<title>.
// If the page doesn't exist, it redirects to the edit form so the user can
// create it (302 Found → /edit/<title>).
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		// Page not found — redirect to edit so it can be created.
		// http.StatusFound == 302. The browser will follow the redirect.
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

// editHandler serves GET /edit/<title>.
// Loads the existing page for editing, or creates an empty one for new pages.
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

// saveHandler serves POST /save/<title>.
// Reads the submitted form body, saves the page to disk, and redirects to view.
func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// ============================================================
// DONE: Add a handler to make the web root redirect to /view/FrontPage
// ============================================================

// rootHandler redirects bare "/" requests to /view/FrontPage.
// This makes the wiki have a proper home page — visiting the server root
// lands on FrontPage rather than a 404.
//
// r.URL.Path == "/" is checked explicitly because net/http's default mux
// routes any path with no more specific handler to "/". Without this check,
// "/nonexistent" would also redirect to FrontPage instead of returning 404.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

// ============================================================
// Entry point
// ============================================================

func main() {
	// DONE: Add a handler to make the web root redirect to /view/FrontPage
	http.HandleFunc("/", rootHandler)

	// Register routes. Each path prefix is paired with a handler wrapped by
	// makeHandler, which handles validation before calling the real handler.
	//
	// net/http uses a simple longest-prefix router:
	//   /view/ matches /view/anything
	//   /edit/ matches /edit/anything
	//   /save/ matches /save/anything
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Println("Starting wiki on http://localhost:8080")

	// http.ListenAndServe blocks until the server stops.
	// It only returns when an error occurs (e.g. port already in use),
	// so wrapping it with log.Fatal prints the error and exits.
	log.Fatal(http.ListenAndServe(":8080", nil))
}
