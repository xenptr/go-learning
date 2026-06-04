# 03-data-access — Tutorial Notes

Source: https://go.dev/doc/tutorial/database-access

---

## What this tutorial covers

| Topic | Key API |
|---|---|
| Connect to MySQL | `mysql.Config`, `sql.Open`, `db.Ping` |
| Query multiple rows | `db.Query` → `rows.Next` / `rows.Scan` / `rows.Err` |
| Query a single row | `db.QueryRow` → `row.Scan` |
| Insert a row | `db.Exec` → `result.LastInsertId` |

---

## Project structure

```
03-data-access/
├── go.mod              ← module example/data-access
├── go.sum              ← dependency checksums (auto-generated)
├── create-tables.sql   ← MySQL bootstrap script
├── main.go             ← full application
└── NOTES.md            ← this file
```

---

## Prerequisites

- MySQL running locally on `127.0.0.1:3306`
- Go installed

---

## Step 1 — Set up the database

Log in to MySQL and run the SQL script:

```bash
mysql -u root -p
```

```sql
mysql> CREATE DATABASE recordings;
mysql> USE recordings;
mysql> source /path/to/03-data-access/create-tables.sql
```

Verify the seed data:

```sql
mysql> SELECT * FROM album;
```

Expected:

```
+----+---------------+----------------+-------+
| id | title         | artist         | price |
+----+---------------+----------------+-------+
|  1 | Blue Train    | John Coltrane  | 56.99 |
|  2 | Giant Steps   | John Coltrane  | 63.99 |
|  3 | Jeru          | Gerry Mulligan | 17.99 |
|  4 | Sarah Vaughan | Sarah Vaughan  | 34.98 |
+----+---------------+----------------+-------+
```

---

## Step 2 — Install the driver dependency

The MySQL driver is already listed in `go.mod`. Just run:

```bash
go mod tidy
```

Or to add it from scratch (as the tutorial does):

```bash
go get .
```

This fetches `github.com/go-sql-driver/mysql` and its transitive dependency
`filippo.io/edwards25519`, writing their checksums into `go.sum`.

---

## Step 3 — Set credentials via environment variables

Never hard-code database passwords in source code.

**Linux / Mac:**

```bash
export DBUSER=root
export DBPASS=yourpassword
```

**Windows:**

```cmd
set DBUSER=root
set DBPASS=yourpassword
```

---

## Step 4 — Run the program

```bash
go run .
```

Expected output:

```
Connected!
Albums found: [{1 Blue Train John Coltrane 56.99} {2 Giant Steps John Coltrane 63.99}]
Album found: {2 Giant Steps John Coltrane 63.99}
ID of added album: 5
```

---

## Key concepts explained

### `database/sql` vs the driver

`database/sql` is the **standard library abstraction** – it defines types like
`*sql.DB`, `*sql.Rows`, and `*sql.Row` and the methods you call in your code.
It knows nothing about MySQL specifically.

`github.com/go-sql-driver/mysql` is the **concrete driver** that speaks the
MySQL wire protocol. It registers itself under the name `"mysql"` during its
`init()` function, which runs when you blank-import it (`_ "..."`).\
After that, `sql.Open("mysql", dsn)` knows which driver to use.

This separation means you can swap databases by changing the driver import and
`sql.Open` name without touching the rest of your query code.

---

### `*sql.DB` is a connection pool, not a single connection

```go
var db *sql.DB  // pool – safe to use from multiple goroutines
```

The pool opens and closes real connections automatically as load changes.
`db.Ping()` forces one real connection immediately so you catch
misconfiguration at startup rather than on the first query.

---

### Parameterised queries prevent SQL injection

```go
// GOOD – value sent separately from SQL text
rows, err := db.Query("SELECT * FROM album WHERE artist = ?", name)

// BAD – never do this
rows, err := db.Query("SELECT * FROM album WHERE artist = '" + name + "'")
```

The `?` placeholder tells the driver to send the SQL statement and the value
in separate protocol frames. The database never tries to parse the value as SQL.

---

### Multi-row vs single-row query

| Situation | Method | Notes |
|---|---|---|
| Query may return 0–N rows | `db.Query` | Returns `*sql.Rows`; must loop + close |
| Query returns at most 1 row | `db.QueryRow` | Returns `*sql.Row`; simpler, no loop |
| No rows returned (INSERT/UPDATE/DELETE) | `db.Exec` | Returns `sql.Result` |

---

### `defer rows.Close()`

Always defer closing rows immediately after checking the `db.Query` error:

```go
rows, err := db.Query(...)
if err != nil {
    return nil, err
}
defer rows.Close()   // ← right here, before any other return paths
```

Forgetting `rows.Close()` leaks the underlying connection back to the pool,
eventually exhausting it under load.

---

### Checking `rows.Err()` after the loop

`rows.Next()` returns `false` in two cases: normal exhaustion AND a mid-stream
error. The only way to tell them apart is to call `rows.Err()` after the loop:

```go
for rows.Next() { ... }
if err := rows.Err(); err != nil {
    return nil, err   // query failed partway through
}
```

---

### `sql.ErrNoRows`

`db.QueryRow(...).Scan(...)` returns `sql.ErrNoRows` when the SELECT matched
zero rows. Treat it as a domain-level "not found" condition rather than a
hard error:

```go
if err == sql.ErrNoRows {
    return alb, fmt.Errorf("albumByID %d: no such album", id)
}
```

---

### `result.LastInsertId()`

After `db.Exec` on an INSERT, `result.LastInsertId()` returns the value the
database assigned to the `AUTO_INCREMENT` primary key column.
Not all databases support this (e.g. PostgreSQL uses `RETURNING` instead),
but MySQL does.

---

## Topics from the broader Accessing Databases guide

Source: https://go.dev/doc/database/

The sections below cover patterns that are **not** in the original step-by-step
tutorial but are documented in the full database access guide. Each one has a
matching example function in `examples.go`.

---

### Opening with a Connector (`sql.OpenDB`)

`sql.Open` accepts a plain DSN string. `sql.OpenDB` accepts a
`driver.Connector`, which lets you pass driver-specific options that a flat
string cannot express (e.g. a custom TLS config or authentication plugin).
Both return `*sql.DB` — the rest of your code is identical.

```go
connector, err := mysql.NewConnector(cfg)   // driver.Connector
db = sql.OpenDB(connector)                  // no DSN string needed
```

See `openWithConnector()` in `examples.go`.

---

### Prepared statements (`sql.Stmt`)

A prepared statement is SQL parsed and stored by the DBMS **once**. Later
executions send only the parameter values — no SQL text re-parsing.

```
db.Prepare(sql)  →  *sql.Stmt
stmt.QueryRow(params...)
stmt.Exec(params...)
stmt.Query(params...)
stmt.Close()   ← always defer this
```

Rules:
- Prepare once, reuse many times. Preparing inside a hot loop defeats the purpose.
- `defer stmt.Close()` releases the server-side handle.
- Inside a transaction use `tx.Prepare` or `tx.Stmt(existingStmt)`.

See `albumByIDPrepared()` in `examples.go`.

---

### Transactions (`sql.Tx`)

A transaction groups operations so they all commit or all roll back atomically.

**Standard pattern:**

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil { ... }
defer tx.Rollback()   // no-op if Commit succeeds

// use tx.ExecContext / tx.QueryRowContext — NOT db.Exec / db.Query
_, err = tx.ExecContext(ctx, "UPDATE ...")
...

if err = tx.Commit(); err != nil { ... }
```

**Critical rules:**
- Always `defer tx.Rollback()` right after `BeginTx`. If `Commit` succeeds,
  `Rollback` becomes a no-op. If anything fails, the rollback fires automatically.
- Use `tx.*` methods, never `db.*` inside a transaction. Calling `db.Exec`
  inside a transaction runs it **outside** the transaction.
- Never mix Go `tx` API with raw SQL `BEGIN`/`COMMIT` statements.

See `CreateOrder()` in `examples.go`.

---

### Context and cancellation

Every `database/sql` method has a `*Context` variant that accepts a
`context.Context`. This lets you:

- Set a **timeout** — cancel a query that runs too long
- Propagate **cancellation** — if an HTTP client disconnects, the DB query stops too

```go
queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()   // always defer — frees the timer even if the query finishes first

rows, err := db.QueryContext(queryCtx, "SELECT * FROM album")
```

If the timeout fires (or the parent context is cancelled), the driver sends a
cancellation to the DBMS and the method returns a `"context deadline exceeded"`
error.

See `QueryWithTimeout()` in `examples.go`.

---

### Nullable column values (`sql.NullString`, etc.)

Scanning a NULL database value into a plain Go type (`string`, `int64`, ...)
causes an error. Use the `sql.Null*` types instead:

```go
var ns sql.NullString
row.Scan(&ns)

if ns.Valid {
    use(ns.String)   // column had a real value
} else {
    use("(default)") // column was NULL
}
```

Available types: `NullString`, `NullInt32`, `NullInt64`, `NullFloat64`,
`NullBool`, `NullTime`.

See `albumTitleByID()` in `examples.go`.

---

### Multiple result sets (`rows.NextResultSet`)

Some drivers let you send multiple SQL statements in one `Query` call. Each
produces its own result set. Use `rows.NextResultSet()` to advance between them.

```go
rows, _ := db.Query("SELECT * FROM album; SELECT * FROM song")
defer rows.Close()

for rows.Next() { /* first result set */ }

rows.NextResultSet()

for rows.Next() { /* second result set */ }

if err := rows.Err(); err != nil { /* covers ALL result sets */ }
```

For MySQL, you must enable multi-statement mode: `cfg.MultiStatements = true`.

See `multiResultDemo()` in `examples.go`.

---

### Connection pool tuning

`*sql.DB` manages a connection pool automatically. The defaults suit most apps,
but high-traffic services may need tuning. Call these once after opening `db`:

| Setter | Default | Purpose |
|---|---|---|
| `SetMaxOpenConns(n)` | unlimited | Cap total open connections |
| `SetMaxIdleConns(n)` | 2 | Keep N idle connections ready |
| `SetConnMaxIdleTime(d)` | unlimited | Close idle connections older than d |
| `SetConnMaxLifetime(d)` | unlimited | Force-close connections older than d |

Use `db.Stats()` to observe current pool behaviour and validate your settings.

See `configurePool()` in `examples.go`.

---

### Dedicated connections (`sql.Conn`)

For sequences of operations that **must** run on the same physical connection
(e.g. connection-scoped temp tables, advisory locks, DDL with implicit
transaction semantics), use `db.Conn(ctx)` to reserve one:

```go
conn, err := db.Conn(ctx)
defer conn.Close()   // returns the connection to the pool

conn.ExecContext(ctx, "CREATE TEMPORARY TABLE ...")
conn.QueryContext(ctx, "SELECT * FROM ...")
```

Prefer `sql.Tx` for transactional work. Use `sql.Conn` only when a transaction
doesn't fit (e.g. DDL that auto-commits, or lock functions that must stay on
the same connection across multiple statements).

---

### SQL injection — quick reference

```go
// SAFE — sql package sends SQL and value in separate frames
db.Query("SELECT * FROM user WHERE id = ?", id)

// UNSAFE — the full string (including id) is sent as SQL text
db.Query(fmt.Sprintf("SELECT * FROM user WHERE id = %s", id))
```

Placeholder syntax varies by driver:
- MySQL / SQLite: `?`
- PostgreSQL (`pq` / `pgx`): `$1`, `$2`, ...
- Oracle: `:name`
