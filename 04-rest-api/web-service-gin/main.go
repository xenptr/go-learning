// Package main is a RESTful web service for a vintage jazz records store.
//
// Source: https://go.dev/doc/tutorial/web-service-gin
//
// This program demonstrates how to build a JSON REST API with Go and the
// Gin web framework. It covers:
//
//   - Defining a data model with JSON struct tags
//   - Routing HTTP methods + paths to handler functions
//   - Returning JSON responses with appropriate HTTP status codes
//   - Parsing JSON from a request body (POST)
//   - Extracting path parameters from the URL (GET /albums/:id)
//
// API surface:
//
//	GET  /albums       → return all albums as JSON
//	POST /albums       → add a new album from JSON body
//	GET  /albums/:id   → return a single album by ID
//
// Run the server:
//
//	go run .
//
// Then test with curl (see NOTES.md for full examples).
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// album represents a single record album in the store.
//
// Struct tags (the `json:"..."` parts) control how the field appears in
// JSON output and how it is matched during JSON input (BindJSON).
// Without tags, Go would use the capitalized field name (e.g. "ID", "Title")
// which is valid but not idiomatic JSON style.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums is the in-memory store for this example.
//
// In a real service you would query a database instead. Because this slice
// lives only in memory, all data is lost when the server stops.
// It is intentionally a package-level variable to keep the tutorial simple;
// production code would pass it through a repository/service layer.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func main() {
	// gin.Default() creates a router with two middleware already attached:
	//   - Logger  — prints each request method, path, status, and latency
	//   - Recovery — catches panics, logs them, and returns 500 to the client
	// For a bare router with no middleware, use gin.New() instead.
	router := gin.Default()

	// Register routes: HTTP method + path + handler function.
	// Note: we pass the function VALUE (getAlbums), not a CALL (getAlbums()).
	// Gin stores the reference and calls it when a matching request arrives.
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID) // :id is a named path parameter
	router.POST("/albums", postAlbums)

	// router.Run starts an http.Server on the given address.
	// Default address is ":8080" if you call Run() with no argument.
	// The call blocks until the server is stopped (Ctrl+C or a signal).
	router.Run("localhost:8080")
}

// getAlbums handles GET /albums.
// It serialises the full albums slice to indented JSON and writes it to the
// response with HTTP 200 OK.
//
// gin.Context is the heart of Gin — it wraps the raw http.Request and
// http.ResponseWriter and adds helpers for JSON, binding, params, etc.
func getAlbums(c *gin.Context) {
	// c.IndentedJSON writes a human-readable (newlines + spaces) JSON body.
	// Use c.JSON for compact JSON (slightly smaller payload, harder to read).
	// The first argument is the HTTP status code.
	c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums handles POST /albums.
// It reads a JSON body from the request, appends the new album to the slice,
// and returns the created album with HTTP 201 Created.
func postAlbums(c *gin.Context) {
	var newAlbum album

	// c.BindJSON decodes the request body into newAlbum.
	// If the body is missing, malformed, or doesn't match the struct fields,
	// BindJSON writes a 400 Bad Request response automatically and returns
	// a non-nil error. We just return early — no need to write another response.
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}

	// Append the decoded album to the in-memory slice.
	albums = append(albums, newAlbum)

	// Respond with 201 Created and echo back the newly added album.
	// 201 (not 200) is the conventional status for a successful resource creation.
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// getAlbumByID handles GET /albums/:id.
// It reads the :id path parameter, searches the albums slice for a match,
// and returns either the album (200 OK) or an error message (404 Not Found).
func getAlbumByID(c *gin.Context) {
	// c.Param("id") returns the value captured by the :id placeholder in
	// the route path. For a request to /albums/2, this returns "2".
	id := c.Param("id")

	// Linear search through the slice.
	// A real service would use a database query or a map for O(1) lookup.
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return // stop immediately after the first match
		}
	}

	// gin.H is a shorthand for map[string]any — useful for ad-hoc JSON objects.
	// Here it produces: {"message": "album not found"}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}
