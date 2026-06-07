// Package main generates random text using a Markov chain algorithm.
//
// Source: https://go.dev/doc/codewalk/markov/
// Based on: The Practice of Programming, Kernighan & Pike (1999)
//
// ── What this codewalk teaches ──────────────────────────────────────────────
//
//   - Method on a non-struct type: (p Prefix) String() and (p Prefix) Shift()
//     defined on []string — Go lets you add methods to ANY named type, not
//     just structs.
//
//   - io.Reader as an interface parameter: Build(r io.Reader) accepts anything
//     that can be read — a file, stdin, a string reader, a network connection.
//     This is Go's composition-over-inheritance approach to polymorphism.
//
//   - The flag package for CLI arguments: -words and -prefix flags parsed
//     automatically from os.Args.
//
//   - bufio.NewReader for efficient buffered reading from any io.Reader.
//
//   - map[string][]string — a map whose values are slices, the core data
//     structure of the Markov chain.
//
// ── How a Markov chain works ─────────────────────────────────────────────────
//
// Given input text, build a table:
//
//   prefix (N words) → list of words that follow it in the input
//
// Example with prefixLen=2 and input "I am not a number I am a free man":
//
//   "" ""        → [I]
//   "" I         → [am]
//   I am         → [not, a]      ← two different words can follow "I am"
//   am not       → [a]
//   not a        → [number]
//   a number     → [I]
//   number I     → [am]
//   am a         → [free]
//   a free       → [man]
//
// To generate text: start with a blank prefix, pick a random suffix, shift
// the prefix (drop first word, append suffix), repeat.
//
// Run:
//
//	echo "your text here" | go run markov/markov.go -words 50 -prefix 2
//	cat somefile.txt | go run markov/markov.go -words 200
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Prefix is a named type for []string.
// Defining it as its own type lets us attach methods to it.
// Without this, we couldn't write (p Prefix) String() or (p Prefix) Shift().
type Prefix []string

// String implements fmt.Stringer and also serves as the map key.
// Joining words with a space gives a consistent, unique string for each prefix.
// e.g. Prefix{"I", "am"}.String() == "I am"
func (p Prefix) String() string {
	return strings.Join(p, " ")
}

// Shift removes the first word from p and appends word at the end.
// This slides the window forward by one word.
//
// copy(p, p[1:]) copies elements p[1], p[2], ... p[n-1] to p[0], p[1], ...
// then we set the last element to the new word.
//
// e.g. p = ["I", "am"], Shift("not") → p = ["am", "not"]
func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

// Chain holds the Markov chain map and the prefix length.
// The map key is a joined prefix string; the value is a slice of words
// that have followed that prefix in the input text.
type Chain struct {
	chain     map[string][]string
	prefixLen int
}

// NewChain creates an empty Chain with the given prefix length.
func NewChain(prefixLen int) *Chain {
	return &Chain{make(map[string][]string), prefixLen}
}

// Build reads text from r and populates the Markov chain.
//
// r io.Reader — accepting an interface (not *os.File directly) makes this
// function usable with any readable source: stdin, files, HTTP responses,
// test strings via strings.NewReader, etc.
//
// The algorithm:
//  1. Start with an all-empty prefix (blank strings).
//  2. Read one word at a time with fmt.Fscan.
//  3. Append the word to the suffix list for the current prefix.
//  4. Shift the prefix forward (drop oldest word, add new word).
func (c *Chain) Build(r io.Reader) {
	br := bufio.NewReader(r) // wrap in bufio for efficient reads
	p := make(Prefix, c.prefixLen)
	for {
		var s string
		if _, err := fmt.Fscan(br, &s); err != nil {
			break // EOF or error → stop reading
		}
		key := p.String()
		// map[key] returns nil if key doesn't exist; append handles nil slice.
		c.chain[key] = append(c.chain[key], s)
		p.Shift(s)
	}
}

// Generate produces at most n words of random text from the chain.
// It starts with a blank prefix and follows random transitions until
// either n words are produced or no continuations exist for the current prefix.
func (c *Chain) Generate(n int) string {
	p := make(Prefix, c.prefixLen) // start blank, same as Build started
	var words []string
	for i := 0; i < n; i++ {
		choices := c.chain[p.String()]
		if len(choices) == 0 {
			break // no suffix known for this prefix — end of chain
		}
		// Pick a random suffix from the available choices.
		next := choices[rand.Intn(len(choices))]
		words = append(words, next)
		p.Shift(next) // advance the prefix window
	}
	return strings.Join(words, " ")
}

func main() {
	// flag.Int registers a CLI flag and returns a pointer to its value.
	// After flag.Parse(), *numWords and *prefixLen hold the parsed values.
	numWords := flag.Int("words", 100, "maximum number of words to print")
	prefixLen := flag.Int("prefix", 2, "prefix length in words")

	flag.Parse()
	rand.Seed(time.Now().UnixNano()) //nolint:staticcheck // fine for this example

	c := NewChain(*prefixLen) // * dereferences the pointer to get the int
	c.Build(os.Stdin)         // read text from stdin to build the chain
	text := c.Generate(*numWords)
	fmt.Println(text)
}
