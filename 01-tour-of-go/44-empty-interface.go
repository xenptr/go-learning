package main

import "fmt"

func main() {
	// Empty interface can store any type.
	var i interface{}

	// (nil, nil)
	describe(i)

	// (42, int)
	i = 42
	describe(i)

	// ("hello", string)
	i = "hello"
	describe(i)
}

func describe(i interface{}) {
	fmt.Printf("(%v, %T)\n", i, i)
}
