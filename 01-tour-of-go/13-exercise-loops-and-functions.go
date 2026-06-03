package main

import (
	"fmt"
)

func Sqrt(x float64) float64 {
	z := 1.0
	// 1st
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
		fmt.Println(z)
	}

	// 2nd
	// for {
	// 	prev := z
	//
	// 	z -= (z*z - x) / (2 * z)
	//
	// 	if math.Abs(z-prev) < 1e-10 {
	// 		break
	// 	}
	// }

	return z
}

func main() {
	fmt.Println(Sqrt(2))
}
