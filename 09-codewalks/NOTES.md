# 09-codewalks — Notes

Sources:
- https://go.dev/doc/codewalk/functions/ (Pig)
- https://go.dev/doc/codewalk/markov/    (Markov)
- https://go.dev/doc/codewalk/sharemem/ (URL poller)

---

## Project structure

```
09-codewalks/
├── go.mod
├── pig/pig.go        ← first-class functions & closures
├── markov/markov.go  ← methods on slice types, io.Reader, flag
├── urlpoll/          ← goroutines + channels as architecture
│   └── urlpoll.go
└── NOTES.md
```

---

## Running the programs

### Pig
```bash
go run pig/pig.go
```
Prints win/loss ratios for each "stay at k" strategy (k=1 to 100).
Look for the peak around k=20-25 — that is the empirically best strategy.

### Markov
```bash
# pipe any text into it
echo "some text here to learn from" | go run markov/markov.go -words 40 -prefix 2

# use a real text file for interesting output
cat /path/to/book.txt | go run markov/markov.go -words 200 -prefix 3
```

### URL poller
```bash
go run urlpoll/urlpoll.go
# Requires internet. Ctrl+C to stop.
# Prints status every 10s. Polls each URL every 60s.
```

---

## Pig — First-class functions

### The core idea

In Go, functions are values. You can:
- Assign a function to a variable: `var f action = roll`
- Store functions in a slice: `strategies := []strategy{stayAtK(5), stayAtK(20)}`
- Pass a function as an argument: `play(stayAtK(10), stayAtK(25))`
- Return a function from a function: `stayAtK(k int) strategy { return func... }`

### Function types

```go
type action   func(score) (score, bool)   // a single move
type strategy func(score) action          // a policy that picks a move
```

`action` and `strategy` are type aliases for function signatures.
Any function with the right signature satisfies the type — no explicit
"implements" declaration needed.

### Closures capture their environment

```go
func stayAtK(k int) strategy {
    return func(s score) action {
        if s.thisTurn >= k { // k is captured from the outer scope
            return stay
        }
        return roll
    }
}
```

`stayAtK(20)` and `stayAtK(25)` each return a different function that
remembers its own `k`. The captured variable lives as long as the closure does.

### Why this matters

Without first-class functions you'd need an interface + struct per strategy:

```go
// Without first-class functions — much more boilerplate
type Strategy interface { Choose(s score) action }
type StayAtK struct { k int }
func (s StayAtK) Choose(sc score) action { ... }
```

Function types give you the same polymorphism with much less ceremony.

---

## Markov — Methods on non-struct types

### Named types get methods

```go
type Prefix []string          // named type based on []string

func (p Prefix) String() string { ... }   // method on Prefix
func (p Prefix) Shift(word string) { ... }
```

Go lets you define methods on ANY named type — not just structs.
`Prefix` is a `[]string` under the hood but has domain-specific behaviour attached.

### `io.Reader` — interface-based I/O

```go
func (c *Chain) Build(r io.Reader) { ... }
```

`io.Reader` is the standard Go interface for "something you can read bytes from":

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

Accepting `io.Reader` instead of `*os.File` means `Build` works with:
- `os.Stdin` — reading from a terminal
- `os.Open("file.txt")` — reading a file
- `strings.NewReader("text")` — reading from a string (great for tests)
- `http.Response.Body` — reading an HTTP response
- Any future type that implements `Read`

This is Go's answer to inheritance: compose behaviour through small interfaces.

### The `flag` package

```go
numWords  := flag.Int("words", 100, "maximum number of words to print")
prefixLen := flag.Int("prefix", 2,   "prefix length in words")
flag.Parse()
// use *numWords, *prefixLen
```

`flag.Int` returns a `*int` (pointer). After `flag.Parse()`, the pointer
points to the parsed value. `--help` is generated automatically.

---

## URL Poller — Share memory by communicating

### The Go concurrency proverb

> "Do not communicate by sharing memory;  
>  instead, share memory by communicating."

**Traditional approach (mutex):**
```go
var mu sync.Mutex
var urlStatus = map[string]string{}

// every goroutine that reads or writes must lock first
mu.Lock()
urlStatus[url] = status
mu.Unlock()
```

**Go approach (channel ownership):**
```go
// urlStatus lives ONLY inside StateMonitor's goroutine
// Other goroutines send State values through a channel
// No mutex needed — only one goroutine ever touches the map
```

### The pipeline architecture

```
pending ──► Poller(s) ──► complete
                │
            status ──► StateMonitor
```

Data flows through channels. Each value is "owned" by one goroutine at a
time — the channel transfer is the handoff of ownership.

### Directional channels enforce data flow

```go
func Poller(in <-chan *Resource, out chan<- *Resource, status chan<- State)
//              ↑ receive-only       ↑ send-only            ↑ send-only
```

The type system prevents Poller from accidentally sending to `in` or
receiving from `out`. Data flow is documented in the function signature
and verified at compile time.

### select — waiting on multiple channels

```go
select {
case <-ticker.C:
    logState(urlStatus)
case s := <-updates:
    urlStatus[s.url] = s.status
}
```

`select` blocks until one of its cases is ready, then executes that case.
If both are ready simultaneously, Go picks one at random.
This is the standard pattern for a goroutine that manages shared state —
it responds to events (timer ticks, incoming updates) without polling.

### Fan-out with multiple goroutines on one channel

```go
for i := 0; i < numPollers; i++ {
    go Poller(pending, complete, status)
}
```

Multiple goroutines reading from the same channel is safe. The channel
guarantees each `*Resource` is delivered to exactly one goroutine.
This is how Go achieves parallelism: launch N workers, feed them through
one shared input channel.

---

## Summary — what each codewalk adds

| Codewalk | New concept vs Tour |
|---|---|
| Pig | Function types, functions as values in slices, closures that return functions |
| Markov | Methods on non-struct types, `io.Reader` as a parameter, `flag` package, `bufio` |
| URL Poller | Channels as ownership transfer (not just synchronisation), directional channel types, `select`, fan-out pattern |
