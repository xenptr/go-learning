// Package main is the entry-point application that consumes the greetings library.
//
// This is the "caller" side of the two-module tutorial from:
// https://go.dev/doc/tutorial/create-module
//
// Key ideas demonstrated here:
//   - Importing a local module via a replace directive in go.mod
//   - Handling the (value, error) return convention
//   - Using the log package for fatal errors (prints message then calls os.Exit(1))
//   - Passing a slice of names to a function that returns a map
//   - Iterating over a map and printing results
//
// In Go, an executable program must be in package main and must define
// a func main() – that is the program's entry point.
package main

import (
	"fmt"
	"log"

	// example.com/greetings is resolved to ../greetings by the replace
	// directive in go.mod. The import path must match the module path
	// declared in greetings/go.mod exactly.
	"example.com/greetings"
)

func main() {
	// ---------- configure the logger ----------
	// log.SetPrefix prepends a label to every log message so you can
	// tell at a glance which program produced the output.
	log.SetPrefix("greetings: ")

	// log.SetFlags(0) removes the default timestamp and source-file info
	// from log output, keeping error messages clean and readable.
	log.SetFlags(0)

	// ---------- single greeting (Hello) ----------
	// greetings.Hello is exported (capital H), so it's visible here.
	// The function returns two values: the message string and an error.
	// We assign both with a single := statement.
	message, err := greetings.Hello("Gladys")
	if err != nil {
		// log.Fatal prints the error message (with our "greetings: " prefix)
		// and then calls os.Exit(1), terminating the program immediately.
		// Use this when an error is unrecoverable.
		log.Fatal(err)
	}
	fmt.Println(message)

	// ---------- multiple greetings (Hellos) ----------
	// A slice literal containing three names.
	// []string{...} declares a slice (dynamic array) of strings.
	names := []string{"Gladys", "Samantha", "Darrin"}

	// greetings.Hellos accepts a []string and returns map[string]string.
	// The map key is the name; the value is the personalised greeting.
	messages, err := greetings.Hellos(names)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println on a map prints it in the form:
	//   map[Darrin:Hail, Darrin! Well met! Gladys:Hi, Gladys. Welcome! ...]
	// Map iteration order in Go is randomised on purpose to prevent
	// programs from accidentally depending on insertion order.
	fmt.Println(messages)

	// ---------- iterate and print each entry neatly ----------
	// range on a map yields (key, value) pairs in random order.
	// This loop gives a cleaner, one-per-line output.
	fmt.Println("\nIndividual greetings:")
	for name, msg := range messages {
		fmt.Printf("  %s → %s\n", name, msg)
	}
}
