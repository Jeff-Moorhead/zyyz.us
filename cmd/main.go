package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"html/template"
	"io"
	"log"
	"net/http"
)

type url struct {
	Root string `form:"root"`
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {

	t := &Template{
		templates: template.Must(template.ParseGlob("static/*.html")),
	}

	e := echo.New()
	e.Renderer = t
	e.GET("/", func(c echo.Context) error {
		log.Print("incoming request!")
		err := c.Render(http.StatusOK, "index", nil)
		if err != nil {
			log.Printf("an unknown error occurred: %v", err)
			return err
		}

		return nil
	})

	e.POST("/add", func(c echo.Context) error {
		var u url
		err := c.Bind(&u)
		if err != nil {
			return c.HTML(http.StatusBadRequest, "<p>could not process url</p>")
		}

		return c.HTML(http.StatusOK, fmt.Sprintf("<p>You submitted %v as the url to shorten</p>", u.Root))
	})

	log.Fatal(e.Start(":8080"))
}
