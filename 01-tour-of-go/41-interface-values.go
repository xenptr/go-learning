package main

import (
	"fmt"
	"math"
)

// Interface requiring M().
type I interface {
	M()
}

type T struct {
	S string
}

// *T implements I.
func (t *T) M() {
	fmt.Println(t.S)
}

type F float64

// F implements I.
func (f F) M() {
	fmt.Println(f)
}

func main() {
	var i I

	// Interface now holds:
	// (value=&T{"Hello"}, type=*T)
	i = &T{"Hello"}
	describe(i)
	i.M()

	// Interface now holds:
	// (value=3.14..., type=F)
	i = F(math.Pi)
	describe(i)
	i.M()
}

// Prints the value and concrete type
// currently stored in the interface.
func describe(i I) {
	fmt.Printf("(%v, %T)\n", i, i)
}
