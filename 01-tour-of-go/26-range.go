package main

import "fmt"

var pow1 = []int{1, 2, 4, 8, 16, 32, 64, 128}

func main() {
	for i, v := range pow1 {
		fmt.Printf("2**%d = %d\n", i, v)
	}

	pow2 := make([]int, 10)
	for i := range pow2 {
		pow2[i] = 1 << uint(i) // == 2**i
	}
	for _, value := range pow2 {
		fmt.Printf("%d\n", value)
	}
}
