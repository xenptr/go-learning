# 04-rest-api — Tutorial Notes

Source: https://go.dev/doc/tutorial/web-service-gin

---

## What this tutorial covers

| Topic | Key API |
|---|---|
| Create a router | `gin.Default()` |
| Register routes | `router.GET`, `router.POST` |
| Return JSON | `c.IndentedJSON`, `c.JSON` |
| Parse a JSON request body | `c.BindJSON` |
| Read URL path parameters | `c.Param` |
| Ad-hoc JSON objects | `gin.H` |

---

## Project structure

```
04-rest-api/
├── go.mod      ← module example/web-service-gin
├── go.sum      ← dependency checksums (auto-generated)
├── main.go     ← full application
└── NOTES.md    ← this file
```

---

## API endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/albums` | Return all albums as JSON |
| `POST` | `/albums` | Add a new album from a JSON body |
| `GET` | `/albums/:id` | Return one album by its ID |

---

## Running the server

```bash
go run .
```

The server listens on `localhost:8080`. Stop it with `Ctrl+C`.

---

## Testing with curl

### GET all albums

```bash
curl http://localhost:8080/albums
```

Expected response (200 OK):

```json
[
    {
        "id": "1",
        "title": "Blue Train",
        "artist": "John Coltrane",
        "price": 56.99
    },
    ...
]
```

### POST a new album

```bash
curl http://localhost:8080/albums \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"id":"4","title":"The Modern Sound of Betty Carter","artist":"Betty Carter","price":49.99}'
```

Expected response (201 Created):

```json
{
    "id": "4",
    "title": "The Modern Sound of Betty Carter",
    "artist": "Betty Carter",
    "price": 49.99
}
```

### GET a single album by ID

```bash
curl http://localhost:8080/albums/2
```

Expected response (200 OK):

```json
{
    "id": "2",
    "title": "Jeru",
    "artist": "Gerry Mulligan",
    "price": 17.99
}
```

Not found response (404):

```json
{
    "message": "album not found"
}
```

---

## Key concepts explained

### Why Gin?

`net/http` is the standard library HTTP package. It is powerful but low-level —
routing by method + path, extracting path parameters, and marshalling JSON all
require manual boilerplate. Gin handles all three out of the box and adds
middleware, validation helpers, and a faster router, without pulling you away
from standard Go idioms.

---

### `gin.Default()` vs `gin.New()`

```go
router := gin.Default()  // Logger + Recovery middleware pre-attached
router := gin.New()      // bare router, no middleware
```

`gin.Default()` is the right choice for development and most production use.
Use `gin.New()` when you want full control over which middleware runs.

---

### Struct tags and JSON serialisation

```go
type album struct {
    ID     string  `json:"id"`     // serialises as "id", not "ID"
    Title  string  `json:"title"`
    Artist string  `json:"artist"`
    Price  float64 `json:"price"`
}
```

Go's `encoding/json` package (used internally by Gin) reads the `json:"..."` tag
to determine the key name in the JSON output. Without tags, exported field names
are used as-is (`"ID"`, `"Title"`, etc.), which is valid but not conventional
JSON style.

Common tag options:
- `json:"name"` — use this key
- `json:"name,omitempty"` — omit the field if it is the zero value
- `json:"-"` — never include this field in JSON

---

### `c.IndentedJSON` vs `c.JSON`

```go
c.IndentedJSON(http.StatusOK, data)  // pretty-printed — easier to read/debug
c.JSON(http.StatusOK, data)          // compact — smaller payload
```

Both set the `Content-Type: application/json` header automatically.
Use `IndentedJSON` during development; switch to `JSON` in production if
payload size matters.

---

### `c.BindJSON` — parsing request bodies

```go
var newAlbum album
if err := c.BindJSON(&newAlbum); err != nil {
    return  // Gin already wrote a 400 Bad Request response
}
```

`BindJSON` reads `c.Request.Body`, decodes JSON into the target struct, and
validates required fields. On failure it automatically writes a `400` response,
so you only need to `return` — do not write another response after a failed bind.

Use `c.ShouldBindJSON` if you want to handle the error yourself (it does NOT
auto-write a 400):

```go
if err := c.ShouldBindJSON(&newAlbum); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}
```

---

### Path parameters with `c.Param`

The colon syntax in the route registers a named capture:

```go
router.GET("/albums/:id", getAlbumByID)
```

Inside the handler, retrieve it by the same name:

```go
id := c.Param("id")   // for /albums/42, returns "42"
```

Gin also supports wildcard parameters (`*action`) that match the rest of the path.

---

### `gin.H` — ad-hoc JSON objects

```go
gin.H{"message": "album not found"}
// equivalent to:
map[string]any{"message": "album not found"}
```

`gin.H` is a type alias for `map[string]any`. It is a concise way to build
one-off JSON objects without defining a struct.

---

### HTTP status codes — conventions

| Situation | Code | Constant |
|---|---|---|
| Successful read | 200 | `http.StatusOK` |
| Resource created | 201 | `http.StatusCreated` |
| Bad request body | 400 | `http.StatusBadRequest` |
| Resource not found | 404 | `http.StatusNotFound` |

Always use `net/http` constants rather than raw integers — they are
self-documenting and impossible to typo.

---

### In-memory data vs a real database

This tutorial stores data in a package-level slice. Consequences:
- All data is lost when the server restarts.
- Not safe for concurrent writes without a mutex or a channel.

For real services, replace the slice with a database layer (see `03-database/`).
The handler signatures and routing code stay exactly the same.
