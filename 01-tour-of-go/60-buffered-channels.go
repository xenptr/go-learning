package main

import "fmt"

func main() {
	// Sends to a buffered channel block only when the buffer is full
	// Receives block when the buffer is empty
	// Buffered channel: can hold 2 values without a receiver
	ch := make(chan int, 2)

	ch <- 1 // buffer: [1]
	ch <- 2 // buffer: [1, 2] (full)
	// ch <- 3 // deadlock: buffer already full

	fmt.Println(<-ch) // remove and receive 1
	fmt.Println(<-ch) // remove and receive 2

	ch <- 3
	fmt.Println(<-ch)
}
