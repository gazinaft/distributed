package util

import (
	"image"
	"image/jpeg"
	"log"
	"os"
)

func GetImageFromFilePath(filepath string) (image.Image, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	image, _, err := image.Decode(f)
	return image, err
}

func WriteImageToFilePath(inImage image.Image, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		log.Printf("failed to create file: %v", err)
		return err
	}
	defer f.Close()
	if err = jpeg.Encode(f, inImage, nil); err != nil {
		log.Printf("failed to encode: %v", err)
		return err
	}
	return nil
}
