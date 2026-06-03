package main

import "fmt"

func main() {
	names := [4]string{
		"John",
		"Paul",
		"George",
		"Ringo",
	}
	fmt.Println(names)

	// a and b do NOT create new arrays.
	// They are slices that point to the same underlying array.
	a := names[0:2] // John Paul
	b := names[1:3] // Paul George

	fmt.Println(a, b)

	// b[0] refers to names[1].
	// Modifying b changes the underlying array.
	b[0] = "XXX"

	// Since a and b share the same array,
	// both slices see the change.
	fmt.Println(a, b)

	// The original array is also modified.
	fmt.Println(names)

	// slice-literals
	// Slice of integers
	q := []int{2, 3, 5, 7, 11, 13}
	fmt.Println(q)

	// Slice of booleans
	r := []bool{true, false, true, true, false, true}
	fmt.Println(r)

	// Slice of anonymous structs
	s := []struct {
		i int
		b bool
	}{
		{2, true},
		{3, false},
		{5, true},
		{7, true},
		{11, false},
		{13, true},
	}
	fmt.Println(s)
}
