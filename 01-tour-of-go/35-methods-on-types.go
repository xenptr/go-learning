package main

import (
	"fmt"
	"math"
)

// MyFloat is a custom type based on float64.
type MyFloat float64

// Abs is a method attached to MyFloat.
func (f MyFloat) Abs() float64 {
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}

func main() {
	f := MyFloat(-math.Sqrt2)

	// Calling a method on a custom type.
	fmt.Println(f.Abs())
}
