package main

import "fmt"

type I interface {
	M()
}

func main() {
	// Nil interface.
	// Contains:
	// (value=nil, type=nil)
	var i I

	describe(i)

	// Panic!
	//
	// Go doesn't know which concrete
	// implementation of M() to call because
	// the interface contains no type.
	i.M()
}

func describe(i I) {
	fmt.Printf("(%v, %T)\n", i, i)
}
