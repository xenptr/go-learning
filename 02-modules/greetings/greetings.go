// Package greetings provides functions for generating greeting messages.
//
// This package is the "library" side of the two-module tutorial from:
// https://go.dev/doc/tutorial/create-module
//
// It demonstrates:
//   - Declaring and exporting functions
//   - Returning multiple values (value + error)
//   - Using slices to hold a set of format strings
//   - Selecting a random element from a slice
//   - Using maps to associate keys with values
//   - Backward-compatible API design (Hello vs Hellos)
package greetings

import (
	"errors"
	"fmt"
	"math/rand"
)

// Hello returns a greeting for the named person.
//
// If name is empty it returns an error; otherwise it picks a random
// greeting format, substitutes the name, and returns the result.
//
// Return signature note: Go functions can return multiple values.
// The convention for functions that may fail is (result, error).
// A nil error means success.
func Hello(name string) (string, error) {
	// Guard clause: reject empty names early.
	// errors.New creates a simple error value from a string message.
	if name == "" {
		// Return the zero value for string ("") together with an error.
		// Returning early keeps the happy-path code un-indented.
		return "", errors.New("empty name")
	}

	// randomFormat() is unexported (lowercase), so it is only callable
	// from inside this package. It returns a format string like "Hi, %v. Welcome!"
	// fmt.Sprintf fills in the %v verb with the value of name.
	message := fmt.Sprintf(randomFormat(), name)

	// nil as the error signals "no problem occurred".
	return message, nil
}

// Hellos returns a map that associates each of the named people
// with a personalised greeting message.
//
// This function was added INSTEAD of changing Hello's signature so that
// existing callers of Hello are not broken – a key principle of
// backward-compatible module design.
//
// Parameter:  names  – a slice of name strings
// Returns:    map[string]string (name → greeting), or an error
func Hellos(names []string) (map[string]string, error) {
	// make(map[key-type]value-type) initialises an empty map.
	// Using make is preferred over a map literal when the map starts empty.
	messages := make(map[string]string)

	// range over the slice; we don't need the numeric index so we
	// discard it with the blank identifier _.
	for _, name := range names {
		// Reuse the existing Hello function – no duplication needed.
		message, err := Hello(name)
		if err != nil {
			// Return nil map + the error so the caller knows what went wrong.
			return nil, err
		}

		// Map assignment: messages["Alice"] = "Hi, Alice. Welcome!"
		messages[name] = message
	}

	return messages, nil
}

// randomFormat returns one of several predefined greeting format strings.
// The returned string contains a single %v verb that the caller fills with a name.
//
// Note: this function is UNEXPORTED (starts with a lowercase letter).
// It is an internal helper – callers outside this package cannot see it.
func randomFormat() string {
	// A slice literal: []string{...}
	// Unlike arrays, slices don't have a fixed size declared in the brackets.
	// Go determines the underlying array size from the number of elements.
	formats := []string{
		"Hi, %v. Welcome!",
		"Great to see you, %v!",
		"Hail, %v! Well met!",
	}

	// rand.Intn(n) returns a random int in [0, n).
	// Using it as an index gives us a uniformly random element.
	return formats[rand.Intn(len(formats))]
}
