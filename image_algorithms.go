package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/crazy3lf/colorconv"
)

const posterization_levels = 8

func posterize(value, levels float64) float64 {
	return math.Round(value*levels) / levels
}

func pixel_posterization(inPixel color.Color, levels float64) (color.Color, error) {
	h, s, v := colorconv.ColorToHSV(inPixel)

	v = posterize(v, levels)

	return colorconv.HSVToColor(h, s, v)
}

func posterize_image(inImage image.Image, levels float64) image.Image {

	newImage := image.NewRGBA(image.Rect(0, 0, inImage.Bounds().Max.X, inImage.Bounds().Max.Y))

	for y := inImage.Bounds().Min.Y; y < inImage.Bounds().Max.Y; y++ {
		for x := inImage.Bounds().Min.X; x < inImage.Bounds().Max.X; x++ {
			new_pixel, err := pixel_posterization(inImage.At(x, y), levels)

			if err == nil {
				newImage.Set(x, y, new_pixel)
			} else {
				fmt.Printf("Failed to posterize a pixel at %d, %d, %v", x, y, err)
				newImage.Set(x, y, inImage.At(x, y))
			}
		}
	}
	return newImage
}
