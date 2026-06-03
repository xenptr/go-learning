package main

import "fmt"

func main() {
	// Zero value of a slice is nil.
	// A nil slice has len=0, cap=0,
	// and does not reference an array.
	var s []int

	fmt.Println(s, len(s), cap(s))

	if s == nil {
		fmt.Println("nil!")
	}
}
