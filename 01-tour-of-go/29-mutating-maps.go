package main

import "fmt"

func main() {
	// Create an empty map of string keys and int values.
	m := make(map[string]int)

	m["Answer"] = 42
	fmt.Println("The value:", m["Answer"])

	m["Answer"] = 48
	fmt.Println("The value:", m["Answer"])

	// Remove the key from the map.
	delete(m, "Answer")

	// When a key doesn't exist:
	// v gets the zero value of the map's value type (0 for int)
	// ok indicates whether the key was found.
	v, ok := m["Answer"]

	fmt.Println("The value:", v, "Present?", ok)
}
