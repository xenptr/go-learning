package main

import "fmt"

// Interface requiring a single method M().
type I interface {
	M()
}

// Struct definition.
type T struct {
	S string
}

// Because T has a method named M(),
// T automatically satisfies interface I.
func (t T) M() {
	fmt.Println(t.S)
}

func main() {
	// Interface variable i.

	// T implements I because it has M().
	// No "implements" keyword required.
	var i I = T{"hello"}

	// Calls T.M()
	i.M()
}
