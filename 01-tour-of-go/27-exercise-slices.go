package main

import "golang.org/x/tour/pic"

func Pic(dx, dy int) [][]uint8 {
	s := make([][]uint8, dy)
	for i := range s {
		s[i] = make([]uint8, dx)
	}

	for y := range s {
		for x := range s[y] {
			s[y][x] = uint8(x * y)
		}
	}
	return s
}

func main() {
	pic.Show(Pic)
}
