package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"log"
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

func writeImageToFilePath(inImage image.Image, filepath, name string) error {
	f, err := os.Create(fmt.Sprintf("%s/%s.jpg", filepath, name))
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

func main() {
	img, err := getImageFromFilePath("./image/dog.jpg")
	if err != nil {
		panic(err)
	}

	writeImageToFilePath(posterize_image(img, 5), "./image", "dog2")

	// IMPORTANT, http get query
	// url := "http://localhost:8080"
	// resp, err := http.Get(url)
	// if err != nil {
	// 	// we will get an error at this stage if the request fails, such as if the
	// 	// requested URL is not found, or if the server is not reachable.
	// 	log.Fatal(err)
	// }
	// defer resp.Body.Close()

	// // if we want to check for a specific status code, we can do so here
	// // for example, a successful request should return a 200 OK status
	// if resp.StatusCode != http.StatusOK {
	// 	// if the status code is not 200, we should log the status code and the
	// 	// status string, then exit with a fatal error
	// 	log.Fatalf("status code error: %d %s", resp.StatusCode, resp.Status)
	// }

	// // print the response
	// data, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(string(data))
}
