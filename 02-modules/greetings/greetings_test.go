// Tests for the greetings package.
//
// Key Go testing conventions used here:
//   - Test files end in _test.go  →  the go test command picks them up automatically.
//   - Test functions are named TestXxx (capital T, then anything).
//   - Each test function receives *testing.T, which provides helpers like
//     t.Errorf (mark failed + log) and t.Fatalf (mark failed + stop test immediately).
//   - Test files belong to the SAME package as the code under test so they can
//     call unexported helpers if needed (though we only use exported ones here).
//
// Run all tests:          go test
// Run with verbose info:  go test -v
package greetings

import (
	"regexp"
	"testing"
)

// TestHelloName verifies that Hello returns a non-error message that
// contains the caller-supplied name somewhere in the output.
//
// We use a regexp instead of an exact string match because randomFormat()
// means the greeting text changes on every run – only the embedded name
// is guaranteed to be present.
func TestHelloName(t *testing.T) {
	name := "Gladys"

	// \b is a word-boundary anchor so "Gladys" won't accidentally match
	// inside a longer word. MustCompile panics if the pattern is invalid,
	// which is fine for test setup code (it would indicate a bug in the test).
	want := regexp.MustCompile(`\b` + name + `\b`)

	msg, err := Hello("Gladys")

	// Two conditions must both be true for the test to pass:
	//   1. the returned message contains the name
	//   2. no error was returned
	// t.Errorf formats a message and marks the test as failed,
	// but continues running (unlike t.Fatalf which stops immediately).
	if !want.MatchString(msg) || err != nil {
		t.Errorf(`Hello("Gladys") = %q, %v, want match for %#q, nil`, msg, err, want)
	}
}

// TestHelloEmpty verifies that Hello rejects an empty name with an error.
//
// This tests the error-handling branch – passing the zero value for string.
// Both conditions must hold: returned message must be "" and err must be non-nil.
func TestHelloEmpty(t *testing.T) {
	msg, err := Hello("")

	// If either condition is wrong the test fails:
	//   - a non-empty message would mean Hello "succeeded" on bad input
	//   - a nil error would mean the error was silently swallowed
	if msg != "" || err == nil {
		t.Errorf(`Hello("") = %q, %v, want "", error`, msg, err)
	}
}
