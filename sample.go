package main

import (
	"image"
	"os"
)

func getImageFromFilePath(filepath string) (image.Image, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	image, _, err := image.Decode(f)
	return image, err
}

func applyKernel(inImage image.Image, kernel []float32)
{
	
}

func main() {

}
