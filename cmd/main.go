// lab 1-4

package main

import (
	"fmt"
	"html/template"
	"io"
	"math/rand"
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

func RandBool() bool {
	return rand.Intn(2) == 1
}

var requestNum int = 0

func HandlePostImage(c echo.Context) error {
	file, err := c.FormFile("image")

	if err != nil {
		return err
	}

	algo := c.FormValue("algo")
	methodString := c.FormValue("methods")

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

	if algo == "1" || algo == "2" {
		// P for Posterize C for convolution
		if requestNum%2 == 0 {
			filename = "P" + filename
		} else {
			filename = "C" + filename
		}
		requestNum++
	}

	if algo == "4" || algo == "3" {
		if len(methodString) == 0 {
			return nil
		}
		filename = methodString + filename
	}

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

	var alteredPic string

	if algo == "1" {
		alteredPic, err = SendImageToServiceSync(filename)
	} else if algo == "2" {
		alteredPic, err = SendImageToServiceAsync(filename)
	} else if algo == "3" {
		alteredPic, err = SendImageToEventStore(filename)
	} else if algo == "4" {
		alteredPic, err = SendImageToOrchestrator(filename)
	} else {
		return echo.ErrBadRequest
	}

	if err != nil {
		return err
	}

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
