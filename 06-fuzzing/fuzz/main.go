// Package main demonstrates Go fuzzing via a string-reversal function.
//
// Source: https://go.dev/doc/tutorial/fuzz
//
// The tutorial walks through two bugs that fuzzing finds automatically:
//
//   Bug 1 — byte-wise reversal corrupts multi-byte UTF-8 characters
//           Fix: reverse by rune, not by byte
//
//   Bug 2 — invalid UTF-8 bytes in the input produce a different invalid
//           UTF-8 sequence after rune-conversion (Go replaces unknown bytes
//           with the replacement character U+FFFD '�')
//           Fix: validate input upfront and return an error for invalid UTF-8
//
// The final Reverse function below is the result after both fixes.
//
// Run normally:
//
//	go run .
//
// Run unit/fuzz tests (seed corpus only, no fuzzing engine):
//
//	go test
//
// Run with the fuzzing engine (generates random inputs):
//
//	go test -fuzz=FuzzReverse
//	go test -fuzz=FuzzReverse -fuzztime 30s   ← stop after 30 s
package main

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

func main() {
	input := "The quick brown fox jumped over the lazy dog"

	// Reverse now returns (string, error) after Bug 2 fix.
	rev, revErr := Reverse(input)
	doubleRev, doubleRevErr := Reverse(rev)

	fmt.Printf("original:       %q\n", input)
	fmt.Printf("reversed:       %q, err: %v\n", rev, revErr)
	fmt.Printf("reversed again: %q, err: %v\n", doubleRev, doubleRevErr)
}

// Reverse returns the string s with its Unicode code points (runes) in
// reverse order, or an error if s is not valid UTF-8.
//
// ── Evolution through the tutorial ──────────────────────────────────────
//
// Version 1 — byte-by-byte reversal (original, BUGGY):
//
//	func Reverse(s string) string {
//	    b := []byte(s)
//	    for i, j := 0, len(b)-1; i < len(b)/2; i, j = i+1, j-1 {
//	        b[i], b[j] = b[j], b[i]
//	    }
//	    return string(b)
//	}
//
//	Problem: multi-byte UTF-8 characters (e.g. '泃' = 3 bytes) get their
//	bytes shuffled individually, producing invalid UTF-8. Fuzzing found
//	this with the input "泃".
//
// Version 2 — rune-by-rune reversal (Bug 1 fixed, Bug 2 still present):
//
//	func Reverse(s string) string {
//	    r := []rune(s)
//	    for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
//	        r[i], r[j] = r[j], r[i]
//	    }
//	    return string(r)
//	}
//
//	Problem: when s contains an invalid UTF-8 byte (e.g. "\x91"), Go's
//	[]rune conversion silently replaces it with U+FFFD ('�'). Reversing
//	twice gives '�' instead of the original '\x91'. Fuzzing found this.
//
// Version 3 — rune-by-rune with upfront UTF-8 validation (both bugs fixed):
func Reverse(s string) (string, error) {
	// Guard: reject input that isn't valid UTF-8 before doing any work.
	// This makes the contract explicit: callers must supply valid UTF-8.
	// Without this check, the rune conversion below silently mutates
	// invalid bytes into '�', breaking the double-reverse property.
	if !utf8.ValidString(s) {
		return s, errors.New("input is not valid UTF-8")
	}

	// []rune(s) decodes the UTF-8 string into a slice of Unicode code points.
	// Each rune is one code point regardless of how many bytes it occupies
	// in UTF-8 (1–4 bytes). This is what makes the reversal Unicode-correct.
	r := []rune(s)

	// Two-pointer swap: walk from both ends toward the middle.
	// i starts at 0 (left), j starts at len-1 (right).
	// Loop runs while i < len/2 so we stop at the midpoint.
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}

	// string(r) re-encodes the rune slice back to UTF-8.
	return string(r), nil
}
