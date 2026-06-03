package main

import (
	"fmt"
	"time"
)

// Custom error type
type MyError struct {
	When time.Time
	What string
}

// Implement the error interface.
//
// Any type that has:
//
//	Error() string
//
// automatically satisfies the built-in error interface.
func (e *MyError) Error() string {
	return fmt.Sprintf(
		"at %v, %s",
		e.When,
		e.What,
	)
}

// Function returns an error
func run() error {
	// Return a custom error value
	return &MyError{
		time.Now(),
		"it didn't work",
	}
}

func main() {
	// Call the function
	if err := run(); err != nil {
		// fmt.Println sees that err implements
		// the error interface and automatically
		// calls err.Error()
		fmt.Println(err)
	}
}
