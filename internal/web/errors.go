package web

import "github.com/labstack/echo/v4"

func RenderError(c echo.Context, err error) error {
	return c.Render(500, "error.tpl.html", errHtmlParams{Err: err})
}

type errHtmlParams struct {
	Err error
}
