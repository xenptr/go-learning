package main

import "fmt"

func main() {
	s := []int{2, 3, 5, 7, 11, 13}
	printSlice(s)

	// Length becomes 0, but the slice still starts
	// at index 0 of the underlying array.
	// Since the start doesn't move, capacity stays 6.
	s = s[:0]
	printSlice(s)

	// Grow the slice back to length 4.
	// Still starts at index 0, so capacity stays 6.
	s = s[:4]
	printSlice(s)

	// Move the slice start forward by 2 positions.
	// Capacity shrinks because fewer elements remain
	// between the new start and the end of the array.
	s = s[2:]
	printSlice(s)
}

func printSlice(s []int) {
	fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
}
