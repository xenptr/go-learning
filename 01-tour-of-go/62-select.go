package main

import "fmt"

// Generates Fibonacci numbers until a quit signal is received.
func fibonacci(c, quit chan int) {
	x, y := 0, 1

	for {
		select {

		// Send next Fibonacci number when a receiver
		// is waiting on channel c.
		case c <- x:
			x, y = y, x+y

		// Stop generation when quit receives a value.
		case <-quit:
			fmt.Println("quit")
			return
		}
	}
}

func main() {
	c := make(chan int)    // Fibonacci numbers
	quit := make(chan int) // Stop signal

	// Consumer goroutine
	go func() {
		// Receive and print first 10 Fibonacci numbers.
		for i := 0; i < 10; i++ {
			fmt.Println(<-c)
		}

		// Tell fibonacci() to stop.
		quit <- 0
	}()

	// Producer
	fibonacci(c, quit)
}
