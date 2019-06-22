package app

import (
	"io"
	"os"
	"path"
	"text/template"

	"github.com/labstack/echo"
	// "github.com/labstack/gommon/log"
)

// Renderer ...
type Renderer struct {
	dir string
}

// Render ...
func (r *Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	t := template.Must(template.ParseFiles(path.Join(r.dir, "layout.html"), path.Join(r.dir, name)))
	err := t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		// log.Errorf("[Render] data:%#+v error:%v", data, err)
	}
	return err
}

func registerTemplates(e *echo.Echo) {
	views := os.Getenv("VIEWS")
	r := &Renderer{
		dir: views,
	}
	e.Renderer = r
}
