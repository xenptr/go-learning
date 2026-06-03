package main

import (
	"fmt"
	"math"
)

// Interface definition.
// Any type that has Abs() float64 satisfies this interface.
type Abser interface {
	Abs() float64
}

func main() {
	var a Abser // Interface variable

	f := MyFloat(-math.Sqrt2)
	v := Vertex{3, 4}

	// MyFloat has Abs() method,
	// therefore it satisfies Abser.
	a = f

	// Abs() is defined on *Vertex,
	// so pointer to Vertex satisfies Abser.
	a = &v

	// Uncommenting this line causes a compile error.
	// Vertex itself does NOT satisfy Abser because
	// Abs() belongs to *Vertex, not Vertex.
	// a = v

	fmt.Println(a.Abs())
}

// Custom float type.
type MyFloat float64

// Method attached to MyFloat.
func (f MyFloat) Abs() float64 {
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}

// Struct definition.
type Vertex struct {
	X, Y float64
}

// Method attached to *Vertex (pointer receiver).
func (v *Vertex) Abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}
