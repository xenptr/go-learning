package main

import (
	"fmt"
	"math"
)

type Vertex struct {
	X, Y float64
}

// Value receiver: receives a copy of Vertex.
func (v Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Pointer receiver: can modify the original Vertex.
func (v *Vertex) Scale(f float64) {
	v.X *= f
	v.Y *= f
}

func main() {
	v := Vertex{3, 4}

	// Modifies the original struct.
	v.Scale(10)

	fmt.Println(v.Abs())
}
