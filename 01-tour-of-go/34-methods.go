package main

import (
	"fmt"
	"math"
)

// Struct representing a point in 2D space
type Vertex struct {
	X, Y float64
}

// Remember: a method is just a function with a receiver argument.
// Method attached to the Vertex type.
// (v Vertex) is called the receiver.
func (v Vertex) Abs() float64 {
	// Calculate distance from origin (0,0)
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// regular function
func AbsFunc(v Vertex) float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func main() {
	v := Vertex{3, 4}

	// Calling a method on a struct value
	fmt.Println(v.Abs())

	fmt.Println(AbsFunc(v))
}
