package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()

	// Sends a value every 100ms.
	tick := time.Tick(100 * time.Millisecond)

	// Sends a single value after 500ms.
	boom := time.After(500 * time.Millisecond)

	// Helper function to show elapsed time.
	elapsed := func() time.Duration {
		return time.Since(start).Round(time.Millisecond)
	}

	for {
		select {

		// Runs whenever the tick channel receives a value.
		case <-tick:
			fmt.Printf("[%6s] tick.\n", elapsed())

		// Runs once after 500ms, then exits.
		case <-boom:
			fmt.Printf("[%6s] BOOM!\n", elapsed())
			return

		// Runs immediately if no channel is ready.
		default:
			fmt.Printf("[%6s]     .\n", elapsed())

			// Prevent busy waiting / high CPU usage.
			time.Sleep(50 * time.Millisecond)
		}
	}
}
