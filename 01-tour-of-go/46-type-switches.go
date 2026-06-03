package main

import "fmt"

// Accepts any type because interface{} can hold anything
func do(i interface{}) {
	// Check the actual type stored inside i
	switch v := i.(type) {

	// If i contains an int
	case int:
		fmt.Printf("Twice %v is %v\n", v, v*2)

	// If i contains a string
	case string:
		fmt.Printf("%q is %v bytes long\n", v, len(v))

	// Runs when no case matches
	default:
		fmt.Printf("I don't know about type %T!\n", v)
	}
}

func main() {
	// i contains an int
	do(21)

	// i contains a string
	do("hello")

	// i contains a bool
	// No matching case, so default runs
	do(true)
}
