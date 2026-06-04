// examples.go — additional database/sql patterns from the Accessing Databases guide.
//
// Source: https://go.dev/doc/database/
//
// This file extends main.go with examples that were NOT in the original
// tutorial (https://go.dev/doc/tutorial/database-access) but ARE covered
// in the broader database access guide. Each example is a standalone
// function you can call from main() or study independently.
//
// Topics covered here:
//   - Opening a handle with sql.OpenDB + a Connector  (topic: Opening a database handle)
//   - Prepared statements                             (topic: Using prepared statements)
//   - Transactions                                    (topic: Executing transactions)
//   - Context / cancellation                          (topic: Canceling in-progress operations)
//   - Nullable column values                          (topic: Querying for data)
//   - Multiple result sets                            (topic: Querying for data)
//   - Connection pool tuning                          (topic: Managing connections)

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)

// ============================================================
// Opening a handle with sql.OpenDB + a Connector
// ============================================================
//
// sql.Open takes a plain DSN string.
// sql.OpenDB takes a driver.Connector, which lets you use driver-specific
// connection features that can't be expressed in a connection string.
//
// Both return *sql.DB – the rest of your code is identical either way.
// Use this pattern when you need fine-grained driver control at connect time.
func openWithConnector(user, pass string) (*sql.DB, error) {
	cfg := mysql.NewConfig()
	cfg.User = user
	cfg.Passwd = pass
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "recordings"

	// mysql.NewConnector builds a driver.Connector from the Config.
	// A Connector can hold state (e.g. a TLS config) that a plain DSN string
	// cannot represent.
	connector, err := mysql.NewConnector(cfg)
	if err != nil {
		return nil, fmt.Errorf("openWithConnector: %v", err)
	}

	// sql.OpenDB accepts any driver.Connector and returns *sql.DB.
	// No error is returned here for the same reason as sql.Open:
	// the pool is initialised but no real connection is made yet.
	handle := sql.OpenDB(connector)
	return handle, nil
}

// ============================================================
// Prepared statements
// ============================================================
//
// A prepared statement is SQL parsed and stored by the DBMS once.
// Subsequent executions send only the parameter values, not the full SQL text.
//
// Benefits:
//   - Slightly faster when the same statement runs many times
//   - Parameterisation is enforced by design – SQL injection is structurally
//     impossible because the SQL and the values travel separately
//
// Lifecycle: Prepare → use repeatedly → Close.
// Always defer stmt.Close() so database resources are freed.

// albumByIDPrepared demonstrates querying a single row via a prepared statement.
// In a real application you would prepare the statement once at startup (or
// in an init function) and reuse the *sql.Stmt across many calls, rather than
// preparing inside the function on every call as shown here for clarity.
func albumByIDPrepared(id int64) (Album, error) {
	// db.Prepare sends the SQL to the DBMS, which parses and stores it.
	// The returned *sql.Stmt holds a reference to that stored statement.
	stmt, err := db.Prepare("SELECT * FROM album WHERE id = ?")
	if err != nil {
		return Album{}, fmt.Errorf("albumByIDPrepared: %v", err)
	}
	// Always close the statement when done. This releases the server-side
	// prepared statement and any associated connection resources.
	defer stmt.Close()

	var alb Album

	// stmt.QueryRow works exactly like db.QueryRow, but you only supply the
	// parameter values (no SQL text – it was already sent during Prepare).
	err = stmt.QueryRow(id).Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			return alb, fmt.Errorf("albumByIDPrepared %d: no such album", id)
		}
		return alb, fmt.Errorf("albumByIDPrepared %d: %v", id, err)
	}

	return alb, nil
}

// ============================================================
// Transactions
// ============================================================
//
// A transaction groups multiple operations so they all succeed or all fail
// together (atomicity). The workflow is always:
//
//   1. db.BeginTx  → get *sql.Tx
//   2. defer tx.Rollback()   (no-op if Commit was already called)
//   3. do work via tx.ExecContext / tx.QueryRowContext / etc.
//   4. tx.Commit() on success
//
// Key rules:
//   - Use tx.Exec / tx.Query etc., NOT db.Exec / db.Query inside a transaction.
//     Calling db.* methods inside a transaction runs them OUTSIDE the transaction.
//   - Do not mix SQL "BEGIN"/"COMMIT" statements with the Go tx API.
//
// The album_order table used here is not in create-tables.sql on purpose –
// this function is meant to be read as a pattern, not executed as-is.
// Add the table if you want to run it:
//
//	CREATE TABLE album_order (
//	  id       INT AUTO_INCREMENT NOT NULL,
//	  album_id INT NOT NULL,
//	  cust_id  INT NOT NULL,
//	  quantity INT NOT NULL,
//	  date     DATETIME NOT NULL,
//	  PRIMARY KEY (id)
//	);
//
// Also requires a `quantity` column on album:
//
//	ALTER TABLE album ADD COLUMN quantity INT NOT NULL DEFAULT 100;

// CreateOrder creates a customer order for an album inside a transaction.
// It checks inventory, decrements it, inserts the order row, and commits –
// all atomically. If any step fails the whole transaction is rolled back.
func CreateOrder(ctx context.Context, albumID, quantity, custID int) (orderID int64, err error) {
	// fail is a small helper that formats errors consistently.
	// Defined as a closure so it can reference the function name.
	fail := func(err error) (int64, error) {
		return 0, fmt.Errorf("CreateOrder: %v", err)
	}

	// Begin a transaction. nil means use the default isolation level.
	// db.BeginTx associates the transaction with ctx so that if ctx is
	// cancelled (e.g. client disconnect), the transaction is rolled back.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fail(err)
	}

	// Defer rollback IMMEDIATELY after BeginTx.
	// If Commit succeeds, Rollback becomes a no-op (sql.Tx tracks this internally).
	// If anything below returns early with an error, Rollback cleans up.
	defer tx.Rollback() //nolint:errcheck // rollback error is irrelevant after a failed tx

	// Step 1: check inventory.
	// We use tx.QueryRowContext (not db.QueryRowContext) so this read happens
	// inside the same transaction as the write below – consistent snapshot.
	var enough bool
	if err = tx.QueryRowContext(ctx,
		"SELECT (quantity >= ?) FROM album WHERE id = ?",
		quantity, albumID,
	).Scan(&enough); err != nil {
		if err == sql.ErrNoRows {
			return fail(fmt.Errorf("no such album"))
		}
		return fail(err)
	}
	if !enough {
		return fail(fmt.Errorf("not enough inventory"))
	}

	// Step 2: decrement inventory.
	// tx.ExecContext keeps this UPDATE inside the transaction; if we crash
	// before Commit, the decrement is automatically rolled back.
	_, err = tx.ExecContext(ctx,
		"UPDATE album SET quantity = quantity - ? WHERE id = ?",
		quantity, albumID,
	)
	if err != nil {
		return fail(err)
	}

	// Step 3: insert the order row.
	result, err := tx.ExecContext(ctx,
		"INSERT INTO album_order (album_id, cust_id, quantity, date) VALUES (?, ?, ?, ?)",
		albumID, custID, quantity, time.Now(),
	)
	if err != nil {
		return fail(err)
	}

	// Retrieve the auto-generated order ID before committing.
	orderID, err = result.LastInsertId()
	if err != nil {
		return fail(err)
	}

	// Step 4: commit. All three operations are applied atomically.
	// After a successful Commit, the deferred Rollback above is a no-op.
	if err = tx.Commit(); err != nil {
		return fail(err)
	}

	return orderID, nil
}

// ============================================================
// Context and cancellation
// ============================================================
//
// Every database/sql method has a *Context twin (QueryContext, ExecContext,
// etc.) that accepts a context.Context. Passing a context lets you:
//
//   - Set a timeout so runaway queries don't block forever
//   - Propagate cancellation from an outer context (e.g. HTTP request context)
//
// The derived context (queryCtx) is automatically cancelled when either:
//   a) the 5-second timeout fires, OR
//   b) the parent ctx is cancelled (e.g. the HTTP client disconnects)
//
// Always defer the cancel function returned by WithTimeout / WithDeadline.
// This frees the timer resources even if the query finishes before the timeout.

// QueryWithTimeout runs a SELECT with a 5-second deadline.
// ctx is the parent context (e.g. an HTTP request context).
func QueryWithTimeout(ctx context.Context) ([]Album, error) {
	// Derive a child context with a 5-second timeout.
	// cancel MUST be called to release the timer, hence the defer.
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Pass queryCtx to QueryContext.
	// If the query takes longer than 5 seconds, the driver cancels it and
	// returns a "context deadline exceeded" error.
	rows, err := db.QueryContext(queryCtx, "SELECT * FROM album")
	if err != nil {
		return nil, fmt.Errorf("QueryWithTimeout: %v", err)
	}
	defer rows.Close()

	var albums []Album
	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			return nil, fmt.Errorf("QueryWithTimeout scan: %v", err)
		}
		albums = append(albums, alb)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("QueryWithTimeout rows: %v", err)
	}

	return albums, nil
}

// ============================================================
// Nullable column values
// ============================================================
//
// When a column is defined as NULLable in the schema, scanning it into a
// plain Go type (string, int64, etc.) will panic or error if the value IS NULL.
// database/sql provides sql.NullString, sql.NullInt64, sql.NullFloat64,
// sql.NullBool, sql.NullTime, etc. to handle this safely.
//
// Each Null* type has two fields:
//   - Valid bool   — true if the database value is NOT NULL
//   - <T>          — the actual value, meaningful only when Valid == true

// albumTitleByID retrieves an album title that might be NULL in the database.
// (In practice our title column is NOT NULL, but this shows the pattern.)
func albumTitleByID(id int64) (string, error) {
	// sql.NullString wraps a string with a Valid flag.
	var ns sql.NullString

	err := db.QueryRow("SELECT title FROM album WHERE id = ?", id).Scan(&ns)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("albumTitleByID %d: not found", id)
		}
		return "", fmt.Errorf("albumTitleByID %d: %v", id, err)
	}

	// Use the Valid flag to decide what to return.
	// If ns.Valid is false, the column held NULL.
	if ns.Valid {
		return ns.String, nil
	}
	// Fall back to a placeholder when the value is NULL.
	return "(untitled)", nil
}

// ============================================================
// Multiple result sets
// ============================================================
//
// Some drivers allow multiple SQL statements in a single Query call.
// Each statement produces its own result set.
// rows.NextResultSet() advances past the current result set to the next one;
// it returns false when there are no more result sets.
//
// Always check rows.Err() ONCE after all result sets are consumed –
// it catches errors from any of them.

// multiResultDemo shows how to iterate two result sets from one query.
// NOTE: MySQL's Go driver supports multi-statement queries when you add
// `multiStatements=true` to the DSN / Config. Without that setting,
// only the first statement's result set is returned.
func multiResultDemo() error {
	// Requires cfg.MultiStatements = true in mysql.Config (or
	// "?multiStatements=true" appended to the DSN string).
	rows, err := db.Query("SELECT id, title FROM album; SELECT id, title FROM album LIMIT 2")
	if err != nil {
		return fmt.Errorf("multiResultDemo: %v", err)
	}
	defer rows.Close()

	fmt.Println("--- Result set 1 (all albums) ---")
	for rows.Next() {
		var id int64
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			return fmt.Errorf("multiResultDemo scan1: %v", err)
		}
		fmt.Printf("  %d: %s\n", id, title)
	}

	// Advance to the second result set.
	// NextResultSet returns false if there is no next set or an error occurred.
	if rows.NextResultSet() {
		fmt.Println("--- Result set 2 (first 2 albums) ---")
		for rows.Next() {
			var id int64
			var title string
			if err := rows.Scan(&id, &title); err != nil {
				return fmt.Errorf("multiResultDemo scan2: %v", err)
			}
			fmt.Printf("  %d: %s\n", id, title)
		}
	}

	// Single rows.Err() call covers errors from ALL result sets.
	if err := rows.Err(); err != nil {
		return fmt.Errorf("multiResultDemo rows: %v", err)
	}

	return nil
}

// ============================================================
// Connection pool tuning
// ============================================================
//
// sql.DB manages a connection pool automatically.
// For most programs the defaults are fine, but high-traffic services may need
// to tune them. Call these setters once after sql.Open / sql.OpenDB.
//
// Use db.Stats() to observe pool behaviour (open connections, wait count, etc.)
// and guide your tuning decisions.

// configurePool demonstrates the four connection-pool knobs.
// Call this once after opening db, before using it.
func configurePool(d *sql.DB) {
	// SetMaxOpenConns: maximum number of open connections (default: unlimited).
	// When all connections are in use, new operations wait.
	// Too low → throughput bottleneck. Too high → overwhelms the DBMS.
	d.SetMaxOpenConns(25)

	// SetMaxIdleConns: maximum idle connections kept in the pool (default: 2).
	// Idle connections can be reused immediately without a round-trip to open a
	// new connection. Raise this when you see frequent connect/disconnect churn.
	d.SetMaxIdleConns(25)

	// SetConnMaxIdleTime: close idle connections older than this duration.
	// Pair with SetMaxIdleConns to reclaim connections during quiet periods
	// after a burst of traffic without keeping them idle forever.
	d.SetConnMaxIdleTime(5 * time.Minute)

	// SetConnMaxLifetime: force-close connections older than this duration,
	// even if they are active. Useful with load-balanced databases where the
	// load-balancer may silently drop long-lived connections.
	d.SetConnMaxLifetime(30 * time.Minute)

	// Log the pool stats — useful for understanding and validating your settings.
	log.Printf("pool stats after configuration: %+v", d.Stats())
}
