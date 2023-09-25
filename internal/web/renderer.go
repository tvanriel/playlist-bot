package web

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/tvanriel/cloudsdk/http"
	"go.uber.org/zap"
)

type Templater struct {
	Templates *template.Template
	Log       *zap.Logger
}

func (t *Templater) Render(w io.Writer, name string, data any, c echo.Context) error {

	return t.Templates.ExecuteTemplate(w, name, data)
}

func NewTemplater(l *zap.Logger) (*Templater, error) {
	t, err := template.ParseGlob("web/templates/*.html")

	if err != nil {
		return nil, err
	}

	s := t.DefinedTemplates()

	l.Info("templates", zap.String("templates", s))

	return &Templater{
		Templates: t,
		Log:       l,
	}, nil
}

func DecorateTemplater(e *http.Http, t *Templater) *http.Http {
	e.Engine.Renderer = t
	return e
}
