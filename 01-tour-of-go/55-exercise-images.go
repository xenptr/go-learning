package main

import (
	"image"
	"image/color"

	"golang.org/x/tour/pic"
)

type Image struct{}

// Return image dimensions
func (i Image) Bounds() image.Rectangle {
	return image.Rect(0, 0, 256, 256)
}

// Return the color model used by this image
func (i Image) ColorModel() color.Model {
	return color.RGBAModel
}

// Return the color of pixel (x, y)
func (i Image) At(x, y int) color.Color {
	v := uint8((x + y) / 2)

	return color.RGBA{
		v,
		v,
		255,
		255,
	}
}

func main() {
	m := Image{}
	pic.ShowImage(m)
}
