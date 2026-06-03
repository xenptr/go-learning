package main

import (
	"fmt"
	"image"
)

func main() {
	// Create a blank RGBA image
	// with bounds (0,0) to (100,100)
	m := image.NewRGBA(
		image.Rect(0, 0, 100, 100),
	)

	// Print image dimensions
	fmt.Println(m.Bounds())

	// Get the color of pixel (0,0)
	// and print its RGBA values
	fmt.Println(m.At(0, 0).RGBA())
}
