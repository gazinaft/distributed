package util

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/crazy3lf/colorconv"
)

func posterize(value, levels float64) float64 {
	return math.Round(value*levels) / levels
}

func pixel_posterization(inPixel color.Color, levels float64) (color.Color, error) {
	h, s, v := colorconv.ColorToHSV(inPixel)

	v = posterize(v, levels)

	return colorconv.HSVToColor(h, s, v)
}

func PosterizeImage(inImage image.Image, levels float64) image.Image {

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

var BaseFilter = [][]float64{
	{0, 0, 0},
	{0, 1, 0},
	{0, 0, 0},
}

var BoxFilter = [][]float64{
	{1.0 / 9, 1.0 / 9, 1.0 / 9},
	{1.0 / 9, 1.0 / 9, 1.0 / 9},
	{1.0 / 9, 1.0 / 9, 1.0 / 9},
}

func ApplyKernel(inImage image.Image, kernel [][]float64) image.Image {

	kernelSize := len(kernel)
	padding := (kernelSize + 1) / 2
	newImage := image.NewRGBA(image.Rect(0, 0, inImage.Bounds().Max.X-padding, inImage.Bounds().Max.Y-padding))

	for y := inImage.Bounds().Min.Y; y < inImage.Bounds().Max.Y-kernelSize; y++ {
		for x := inImage.Bounds().Min.X; x < inImage.Bounds().Max.X-kernelSize; x++ {

			r := 0.0
			g := 0.0
			b := 0.0
			a := 0.0

			for xk := 0; xk < kernelSize; xk++ {
				for yk := 0; yk < kernelSize; yk++ {
					oldPixel := inImage.At(x+xk, y+yk)

					oldR, oldG, oldB, oldA := oldPixel.RGBA()

					r += kernel[xk][yk] * float64(oldR)
					g += kernel[xk][yk] * float64(oldG)
					b += kernel[xk][yk] * float64(oldB)
					a += kernel[xk][yk] * float64(oldA)
				}
			}

			intR := uint8(r)
			intG := uint8(g)
			intB := uint8(b)
			intA := uint8(a)

			newPixel := color.RGBA{intR, intG, intB, intA}

			newImage.Set(x, y, newPixel)
		}
	}
	return newImage
}
