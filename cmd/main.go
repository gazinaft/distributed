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

func SendImageToService(c echo.Context) error {
	file, err := c.FormFile("image")

	if err != nil {
		return err
	}
	fmt.Printf("Successfully submitted form")

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

	htmlToImageFilePath := fmt.Sprintf("<img src=\"/images/%s\" id=\"returned-image\">", filename)
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

	e.POST("/upload", SendImageToService)

	e.Logger.Fatal(e.Start(":8080"))
}
