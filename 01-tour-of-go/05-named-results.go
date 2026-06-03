package main

import "fmt"

func split(sum int) (x, y int) {
	// x and y are declared as return variables.
	x = sum * 4 / 9
	y = sum - x

	// Naked return: returns the current values of x and y.
	return
}

func main() {
	fmt.Println(split(17))
}
