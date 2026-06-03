package main

import "fmt"

func main() {
	fmt.Println("counting")

	// defer schedules a function call to run
	// when the surrounding function returns.
	for i := 0; i < 10; i++ {
		// Each call is pushed onto a stack.
		// Arguments are evaluated immediately,
		// but execution happens later.
		defer fmt.Println(i)
	}

	fmt.Println("done")

	// When main() ends, deferred calls run
	// in reverse order:
	// 9 8 7 6 5 4 3 2 1 0
}
