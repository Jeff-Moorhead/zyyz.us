package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

const (
	Schema = `
CREATE TABLE IF NOT EXISTS links (
    root TEXT NOT NULL,
    shortened TEXT NOT NULL,
    CONSTRAINT pk_root_shortened PRIMARY KEY (root, shortened)
)
`

	CreateLink = `INSERT INTO links VALUES ($1, $2)`

	GetRoot = `SELECT root FROM links WHERE shortened = $1`
)

type entry struct {
	Root string `form:"root" db:"root"`
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func createUniqueId() string {

	// We'll use UUIDs just to be sure we don't get a collision
	// before the heat death of the universe
	id := uuid.New()
	return id.String()
}

// TODO: refactor this stuff
func main() {

	t := &Template{
		templates: template.Must(template.ParseGlob("static/*.html")),
	}

	dbconn, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("could not connect to the database: %v", err)
	}

	dbconn.MustExec(Schema)

	e := echo.New()
	e.Use(middleware.HTTPSWWWRedirect())
	e.Renderer = t
	e.GET("/", func(c echo.Context) error {
		log.Print("incoming request!")
		err := c.Render(http.StatusOK, "index", nil)
		if err != nil {
			log.Printf("an unknown error occurred: %v", err)
			return c.HTML(http.StatusInternalServerError, "<p>an unknown error occurred</p>")
		}

		return nil
	})

	e.POST("/add", func(c echo.Context) error {
		var u entry
		err := c.Bind(&u)
		if err != nil {
			log.Printf("could not process incoming entry: %v", err)
			return c.HTML(http.StatusBadRequest, "<p>could not process entry</p>")
		}

		id := createUniqueId()
		shortened := fmt.Sprintf("https://zyyz.us/%v", id)
		dbconn.MustExec(CreateLink, u.Root, id)
		return c.HTML(http.StatusOK, fmt.Sprintf("<p>Shortened link: <a href=%v target=_blank>%v</a></p>", shortened, shortened))
	})

	e.GET("/:shortened", func(c echo.Context) error {
		id := c.Param("shortened")
		var uri entry
		err := dbconn.Get(&uri, GetRoot, id)
		if err != nil {
			log.Printf("could not get root for id %v: %v", id, err)
			return c.HTML(http.StatusBadRequest, fmt.Sprintf("<p>could not find entry for %v</p>", id))
		}

		u, err := url.Parse(uri.Root)
		if err != nil {
			log.Printf("could not parse url %v: %v", uri.Root, err)
			return c.HTML(http.StatusInternalServerError, fmt.Sprint(`<p class="error">could not parse url %v</p>`, uri.Root))
		}

		if u.Scheme == "" {
			u.Scheme = "https" // assume https for missing schemes
		}

		return c.Redirect(http.StatusMovedPermanently, u.String())
	})

	log.Fatal(e.Start(":8080"))
}
