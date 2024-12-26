package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

type ImageLink struct {
	ImageLink string
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

func SendImageToService(filename string) (string, error) {
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

func HandlePostImage(c echo.Context) error {
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
	fullFilePath := fmt.Sprintf("./images/%s", filename)
	dst, err := os.Create(fullFilePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	alteredPic, err := SendImageToService(filename)

	htmlToImageFilePath := fmt.Sprintf("<img src=\"/images/%s\" id=\"returned-image\">", alteredPic)
	fmt.Println(htmlToImageFilePath)

	return c.HTML(http.StatusOK, htmlToImageFilePath)
}

func main() {

	e := echo.New()
	e.Use(middleware.Logger())
	e.Static("/images", "images")
	e.Static("/css", "css")

	e.Renderer = newTemplate()

	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", nil)
	})

	e.POST("/upload", HandlePostImage)

	e.Logger.Fatal(e.Start(":8080"))
}
