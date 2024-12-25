package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func getUser(c echo.Context) error {
	id := "5"
	return c.String(http.StatusOK, id)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.GET("/", getUser)

	e.Logger.Fatal(e.Start(":8080"))

}
