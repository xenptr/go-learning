package main

import (
	"fmt"
	"time"
)

// Function that prints a string 5 times
func say(s string) {
	for i := 0; i < 5; i++ {
		// Wait for 100 milliseconds
		time.Sleep(100 * time.Millisecond)

		// Print the given string
		fmt.Println(s)
	}
}

func main() {
	// Start a new goroutine (runs concurrently)
	go say("world")

	// Run in the main goroutine
	// main() waits here until all 5 "hello" prints are done
	say("hello")

	// After say("hello") finishes, main() returns
	// Program exits immediately
	// Any unfinished goroutines are stopped
}
