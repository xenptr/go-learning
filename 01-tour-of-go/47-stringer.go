package main

import "fmt"

// Define a custom type
type Person struct {
	Name string
	Age  int
}

// Implement the Stringer interface
// This tells Go how to convert Person to a string
func (p Person) String() string {
	return fmt.Sprintf("%v (%v years)", p.Name, p.Age)
}

func main() {
	// Create two Person values
	a := Person{"Arthur Dent", 42}
	z := Person{"Zaphod Beeblebrox", 9001}

	// fmt.Println automatically calls String()
	fmt.Println(a, z)
}
