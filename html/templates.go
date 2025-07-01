package html

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/controllers"
	"github.com/sduoduo233/pbb/db"
	"github.com/sduoduo233/pbb/update"
)

//go:embed html/*
var html embed.FS

type TemplateRenderer struct {
	Template interface {
		ExecuteTemplate(wr io.Writer, name string, data any) error
	}
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// add csrf token to context
	dataMap, ok := data.(controllers.D)
	if !ok {
		return fmt.Errorf("render template: data must have type (controllers.D)")
	}
	token, ok := c.Get("CSRF_TOKEN").(string)
	if !ok {
		return fmt.Errorf("render template: csrf token must not be empty")
	}
	dataMap["csrf"] = token

	// return form value for given key
	if c.Request().Method == "POST" {
		c.Request().ParseForm()
		m := make(map[string]string)
		for k, v := range c.Request().PostForm {
			m[k] = v[0]
		}
		dataMap["post"] = m
	}

	// login state
	dataMap["login"] = controllers.GetUser(c) != nil

	// site name
	var siteName string
	err := db.DB.Get(&siteName, "SELECT value FROM settings WHERE key = 'site_name'")
	if err != nil {
		return fmt.Errorf("render template: get site name: %w", err)
	}
	dataMap["site_name"] = siteName

	// version
	dataMap["version"] = update.CURRENT_VERSION

	return t.Template.ExecuteTemplate(w, name, dataMap)
}

func LoadTemplates() *template.Template {
	t, err := template.ParseFS(html, "html/**")
	if err != nil {
		slog.Error("parse template", "err", err)
		panic(err)
	}
	return t
}
