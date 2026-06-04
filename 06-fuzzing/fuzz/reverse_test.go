// reverse_test.go — unit test + fuzz test for Reverse.
//
// Both test types live in the same file, which is normal in Go.
// The go test command automatically picks them up:
//
//	go test              → runs TestReverse (unit) + FuzzReverse (seed corpus only)
//	go test -fuzz=FuzzReverse  → also runs the fuzzing engine
//
// ── Test function naming conventions ────────────────────────────────────
//
//	TestXxx   (*testing.T)   — regular unit test
//	BenchmarkXxx (*testing.B) — benchmark
//	FuzzXxx   (*testing.F)   — fuzz test  (Go 1.18+)
package main

import (
	"testing"
	"unicode/utf8"
)

// TestReverse is a table-driven unit test.
// It checks specific known inputs against their expected reversed output.
// Unit tests are fast and deterministic — good for documenting known behaviour.
func TestReverse(t *testing.T) {
	testcases := []struct {
		in, want string
	}{
		{"Hello, world", "dlrow ,olleH"},
		{" ", " "},    // single space: reversing a single char is a no-op
		{"!12345", "54321!"},
	}

	for _, tc := range testcases {
		// Reverse now returns (string, error); valid ASCII never errors.
		rev, err := Reverse(tc.in)
		if err != nil {
			t.Errorf("Reverse(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if rev != tc.want {
			t.Errorf("Reverse(%q) = %q, want %q", tc.in, rev, tc.want)
		}
	}
}

// FuzzReverse is a fuzz test for the Reverse function.
//
// Key differences from a unit test:
//
//   - Signature: FuzzXxx(f *testing.F)  not  TestXxx(t *testing.T)
//   - f.Add(...)       adds values to the seed corpus (known interesting inputs)
//   - f.Fuzz(func(t, orig)) is the fuzz target — the engine calls this with
//     generated inputs. The inner function receives *testing.T and the fuzz
//     arguments (one string here).
//
// When run WITHOUT -fuzz:
//   The engine only runs the seed corpus — essentially a unit test. Fast, deterministic.
//
// When run WITH -fuzz=FuzzReverse:
//   The engine mutates and generates inputs, running them against the fuzz target
//   until it finds a failure or you stop it. Failures are saved to
//   testdata/fuzz/FuzzReverse/ and replayed automatically on future `go test` runs.
//
// Properties we can verify without knowing the expected output:
//
//  1. Double-reverse idempotency: Reverse(Reverse(s)) == s
//     A string reversed twice must equal the original. If this fails,
//     the reversal is destroying information.
//
//  2. UTF-8 preservation: if the input is valid UTF-8, the output must also be.
//     A correct Unicode-aware reverse must not corrupt the encoding.
func FuzzReverse(f *testing.F) {
	// Seed corpus — hand-picked interesting inputs the engine starts from.
	// f.Add can take multiple arguments if the fuzz target takes multiple params.
	// Here Reverse takes one string, so f.Add takes one string.
	testcases := []string{"Hello, world", " ", "!12345"}
	for _, tc := range testcases {
		f.Add(tc)
	}

	// Fuzz target. The engine calls this function repeatedly with mutated inputs.
	// orig is the engine-generated string for this iteration.
	f.Fuzz(func(t *testing.T, orig string) {
		// Attempt to reverse. Skip this input if it's invalid UTF-8 —
		// that is a documented precondition of Reverse, not a bug.
		rev, err1 := Reverse(orig)
		if err1 != nil {
			return // t.Skip() is an alternative; return is simpler
		}

		// Reverse the reversed string. This must also succeed (a valid
		// UTF-8 output, reversed again, is still valid UTF-8).
		doubleRev, err2 := Reverse(rev)
		if err2 != nil {
			return
		}

		// Property 1: double-reverse idempotency.
		if orig != doubleRev {
			t.Errorf("double-reverse mismatch: orig=%q, doubleRev=%q", orig, doubleRev)
		}

		// Property 2: UTF-8 preservation.
		// We only check this when orig is valid UTF-8, because we already
		// return early when orig is invalid. This is a belt-and-suspenders check.
		if utf8.ValidString(orig) && !utf8.ValidString(rev) {
			t.Errorf("Reverse produced invalid UTF-8: input=%q, output=%q", orig, rev)
		}
	})
}
