// Package main is the entry point for the data-access tutorial application.
//
// Source: https://go.dev/doc/tutorial/database-access
//
// This program demonstrates how to use the standard library's database/sql
// package together with a third-party MySQL driver to:
//
//   - Connect to a MySQL database
//   - Query for multiple rows  (DB.Query  → rows loop)
//   - Query for a single row   (DB.QueryRow)
//   - Insert a new row         (DB.Exec)
//
// Prerequisites:
//   - MySQL running locally on 127.0.0.1:3306
//   - A "recordings" database created and seeded via create-tables.sql
//   - Environment variables DBUSER and DBPASS set to your MySQL credentials
//
// Run:
//
//	export DBUSER=root
//	export DBPASS=yourpassword
//	go run .
package main

import (
	"database/sql" // standard library package for SQL databases
	"fmt"
	"log"
	"os"

	// The underscore import is a "blank import":
	// we don't call any of its symbols directly, but importing it causes the
	// package's init() function to run, which registers the "mysql" driver
	// with database/sql's internal driver registry.
	// After this registration, sql.Open("mysql", ...) knows what to do.
	_ "github.com/go-sql-driver/mysql"

	// mysql (non-blank) is used only to build the DSN via mysql.Config.
	// Separating DSN construction from sql.Open keeps the code readable.
	"github.com/go-sql-driver/mysql"
)

// db is a package-level variable that holds the database handle.
//
// *sql.DB is NOT a single connection – it is a connection pool managed by
// the database/sql package. It is safe to use concurrently from multiple
// goroutines. In production code you would typically pass db as a parameter
// or embed it in a struct rather than using a global, but a global keeps
// this tutorial example straightforward.
var db *sql.DB

// Album maps to a row in the album table.
// Field names are capitalised (exported) so they are visible outside this
// package if needed. Their types must be compatible with the MySQL column
// types: INT → int64, VARCHAR → string, DECIMAL → float32.
type Album struct {
	ID     int64
	Title  string
	Artist string
	Price  float32
}

func main() {
	// ------------------------------------------------------------------ //
	// 1. Capture connection properties
	// ------------------------------------------------------------------ //
	// mysql.NewConfig() returns a Config with sensible defaults already set
	// (e.g. the default charset, timeout, etc.).
	// We then override the fields we care about.
	cfg := mysql.NewConfig()

	// Read credentials from environment variables so we never hard-code
	// secrets in source code. Set them before running:
	//   export DBUSER=root
	//   export DBPASS=yourpassword
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")

	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306" // host:port of the MySQL server
	cfg.DBName = "recordings"   // the database we created with create-tables.sql

	// ------------------------------------------------------------------ //
	// 2. Get a database handle
	// ------------------------------------------------------------------ //
	// sql.Open does NOT necessarily open a real connection right away –
	// it validates the driver name and DSN format and prepares the pool.
	// The actual connection happens lazily (first query) or explicitly via Ping.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		// log.Fatal prints the message and calls os.Exit(1).
		// Appropriate here: without a valid DB handle nothing else can work.
		log.Fatal(err)
	}

	// ------------------------------------------------------------------ //
	// 3. Verify the connection with Ping
	// ------------------------------------------------------------------ //
	// DB.Ping forces the pool to open a real connection and round-trip to
	// the server. This is the earliest point where a wrong password,
	// unreachable host, or missing database would be detected.
	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	// ------------------------------------------------------------------ //
	// 4. Query for multiple rows
	// ------------------------------------------------------------------ //
	albums, err := albumsByArtist("John Coltrane")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Albums found: %v\n", albums)

	// ------------------------------------------------------------------ //
	// 5. Query for a single row
	// ------------------------------------------------------------------ //
	// Hard-code ID 2 to test the single-row query path.
	alb, err := albumByID(2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Album found: %v\n", alb)

	// ------------------------------------------------------------------ //
	// 6. Insert a new row
	// ------------------------------------------------------------------ //
	albID, err := addAlbum(Album{
		Title:  "The Modern Sound of Betty Carter",
		Artist: "Betty Carter",
		Price:  49.99,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID of added album: %v\n", albID)
}

// albumsByArtist queries for all albums whose artist column matches name.
//
// It demonstrates the multi-row query pattern:
//
//	DB.Query → rows loop with rows.Next() → rows.Scan → rows.Err check
//
// The ? placeholder prevents SQL injection: the driver sends the SQL text
// and the parameter value separately, so the value is never interpreted as
// SQL syntax regardless of what it contains.
func albumsByArtist(name string) ([]Album, error) {
	// Declare a nil slice; append will allocate as needed.
	var albums []Album

	// DB.Query executes the SELECT and returns *sql.Rows.
	// The second argument fills the ? placeholder – never use fmt.Sprintf
	// to build SQL strings, as that opens the door to SQL injection.
	rows, err := db.Query("SELECT * FROM album WHERE artist = ?", name)
	if err != nil {
		// Wrap the error with context so the caller can tell which function failed.
		return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
	}
	// defer rows.Close() ensures the underlying connection is returned to the
	// pool when this function exits, even if we return early on error.
	defer rows.Close()

	// rows.Next() advances to the next row. It returns false when there are
	// no more rows OR when an iteration error occurs.
	for rows.Next() {
		var alb Album

		// rows.Scan copies the current row's column values into the provided
		// pointers in left-to-right column order (matching the SELECT * columns:
		// id, title, artist, price).
		// We pass pointers using the & address-of operator.
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
		}

		// append grows the slice by one element.
		albums = append(albums, alb)
	}

	// rows.Next() returns false both when exhausted AND when an error occurs.
	// We must check rows.Err() after the loop to distinguish the two cases.
	// If we skip this check we could return a partial (wrong) result silently.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("albumsByArtist %q: %v", name, err)
	}

	return albums, nil
}

// albumByID queries for the album whose id column matches the given id.
//
// It demonstrates the single-row query pattern using DB.QueryRow.
// QueryRow is simpler than Query when you know at most one row will come back:
// there is no rows loop, and the error is deferred to Scan rather than
// returned from QueryRow itself.
func albumByID(id int64) (Album, error) {
	var alb Album

	// DB.QueryRow returns *sql.Row (not *sql.Rows).
	// It never returns an error directly; any query error surfaces in Scan.
	row := db.QueryRow("SELECT * FROM album WHERE id = ?", id)

	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
		// sql.ErrNoRows is the sentinel value database/sql uses when the query
		// matched zero rows. We convert it into a friendlier message.
		if err == sql.ErrNoRows {
			return alb, fmt.Errorf("albumByID %d: no such album", id)
		}
		// Any other scan error (type mismatch, connection lost, etc.)
		return alb, fmt.Errorf("albumByID %d: %v", id, err)
	}

	return alb, nil
}

// addAlbum inserts a new album row into the database and returns the
// auto-generated primary key id.
//
// It demonstrates DB.Exec, which is used for SQL statements that modify
// data (INSERT, UPDATE, DELETE) and do not return rows.
// DB.Exec returns a sql.Result that carries LastInsertId and RowsAffected.
func addAlbum(alb Album) (int64, error) {
	// DB.Exec sends the INSERT statement.
	// Three ? placeholders correspond to the three values that follow.
	// We intentionally omit the id column because AUTO_INCREMENT fills it.
	result, err := db.Exec(
		"INSERT INTO album (title, artist, price) VALUES (?, ?, ?)",
		alb.Title, alb.Artist, alb.Price,
	)
	if err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}

	// Result.LastInsertId returns the id assigned by AUTO_INCREMENT.
	// Not all databases support this; MySQL does.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}

	return id, nil
}
