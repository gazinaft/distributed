package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gazinaft/distributed/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func modifyImage(c echo.Context) error {

	filename := c.Param("ImagePath")

	fmt.Printf("ImagePath of original image %s \n", filename)

	img, err := util.GetImageFromFilePath(fmt.Sprintf("../images/%s", filename))
	if err != nil {
		panic(err)
	}

	uuid, err := uuid.NewRandom()

	if err != nil {
		return err
	}

	// get unique filename
	newFilename := uuid.String() + filepath.Ext(filename)
	fmt.Printf("created uuid %s \n", newFilename)

	err = util.WriteImageToFilePath(util.PosterizeImage(img, 5), newFilename)

	if err != nil {
		return err
	}

	return c.String(http.StatusOK, newFilename)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/", modifyImage)

	e.Logger.Fatal(e.Start(":8081"))

}
