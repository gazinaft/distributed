package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func SendImageToServiceSync(filename string) (string, error) {
	url := "http://localhost:8081/"

	req, err := http.NewRequest(http.MethodGet, url, nil)

	query := req.URL.Query()
	query.Add("ImagePath", filename)
	req.URL.RawQuery = query.Encode() // set query param
	fmt.Printf("sent filename %s to url \n", filename)

	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("Service returned %s \n", string(data))

	return string(data), nil
}

func HandlePostImageSync(c echo.Context) error {
	file, err := c.FormFile("image")

	if err != nil {
		return err
	}
	fmt.Println("Successfully submitted form")

	src, err := file.Open()
	if err != nil {
		return err
	}
	fmt.Println("Successfully opened file")

	defer src.Close()

	uuid, err := uuid.NewRandom()

	if err != nil {
		return err
	}

	// get unique filename
	filename := uuid.String() + filepath.Ext(file.Filename)
	fmt.Printf("created uuid %s \n", filename)

	// Destination
	fullFilePath := fmt.Sprintf("images/%s", filename)
	dst, err := os.Create(fullFilePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	alteredPic, err := SendImageToServiceSync(filename)

	if err != nil {
		return err
	}

	htmlToImageFilePath := fmt.Sprintf("<img src=\"/images/%s\" id=\"returned-image\">", alteredPic)
	fmt.Println(htmlToImageFilePath)

	return c.HTML(http.StatusOK, htmlToImageFilePath)
}
