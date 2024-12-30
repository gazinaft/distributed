// lab 1

package main

import (
	"fmt"
	"io"
	"net/http"
)

func SendImageToServiceSync(filename string) (string, error) {
	url := "http://image-nginx-lb/"

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
