package main

import "fmt"

var i, j int = 1, 2

func main() {
	// Types are inferred as bool, bool, and string.
	// := is shorthand for var with inferred type (inside functions only).
	c, python, java := true, false, "no!"

	fmt.Println(i, j, c, python, java)
}
