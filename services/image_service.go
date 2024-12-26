package services

import (
	"net/http"

	"github.com/gazinaft/distributed/util"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func modifyImage(c echo.Context) error {

	filename := c.Param("")

	img, err := util.GetImageFromFilePath(filename)
	if err != nil {
		panic(err)
	}

	util.WriteImageToFilePath(util.PosterizeImage(img, 5), "./image", "dog2")

	id := "5"
	return c.String(http.StatusOK, id)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/", modifyImage)

	e.Logger.Fatal(e.Start(":8080"))

}
