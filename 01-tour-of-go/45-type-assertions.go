package main

import "fmt"

func main() {
	// interface{} can hold a value of any type
	var i interface{} = "hello"

	// Assert that i contains a string
	// If not, the program will panic
	s := i.(string)
	fmt.Println(s)

	// Safe type assertion
	// ok is true if i contains a string
	s, ok := i.(string)
	fmt.Println(s, ok)

	// Check whether i contains a float64
	// Since i contains a string, ok will be false
	f, ok := i.(float64)
	fmt.Println(f, ok)

	// Unsafe assertion
	// This will panic because i contains a string,
	// not a float64
	f = i.(float64)
	fmt.Println(f)
}
