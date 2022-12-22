package main

import (
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

func main() {

	e := echo.New()
	e.GET("/home", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "<p>Hello world!</p>")
	})

	log.Fatal(e.Start(":8080"))
}
