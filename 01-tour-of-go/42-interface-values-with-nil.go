package main

import "fmt"

type I interface {
	M()
}

type T struct {
	S string
}

// Method can safely handle a nil receiver.
func (t *T) M() {
	// t may be nil when called through an interface.
	if t == nil {
		fmt.Println("<nil>")
		return
	}

	fmt.Println(t.S)
}

func main() {
	var i I

	// Nil pointer of type *T.
	var t *T

	// Interface now contains:
	// (value=nil, type=*T)
	i = t

	describe(i)

	// Calls (*T).M() with t == nil.
	i.M()

	// Interface now contains:
	// (value=&T{"hello"}, type=*T)
	i = &T{"hello"}

	describe(i)
	i.M()
}

func describe(i I) {
	fmt.Printf("(%v, %T)\n", i, i)
}
