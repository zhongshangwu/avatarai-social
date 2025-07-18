package api

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplateRenderer(pattern string) *TemplateRenderer {
	return &TemplateRenderer{
		templates: template.Must(template.ParseGlob(pattern)),
	}
}

type LoginPageData struct {
	User     *LoginPageUserData
	Messages []string
	URLs     map[string]string
}

type LoginPageUserData struct {
	Handle string
}
