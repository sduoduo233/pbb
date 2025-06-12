package main

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sduoduo233/pbb/controllers"
	"github.com/sduoduo233/pbb/db"
	"github.com/sduoduo233/pbb/html"
)

func main() {

	// database
	db.Init()

	// echo
	e := echo.New()

	// logger
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				slog.Info("request", "method", v.Method, "uri", v.URI, "status", v.Status)
			} else {
				slog.Info("request", "method", v.Method, "uri", v.URI, "status", v.Status, "error", v.Error)
			}
			return nil
		},
		LogMethod: true,
		LogURI:    true,
		LogStatus: true,
		LogError:  true,
	}))

	// template
	e.Renderer = &html.TemplateRenderer{
		Template: html.LoadTemplates(),
	}

	// controllers
	controllers.Route(e.Group(""))

	err := e.Start(":3005")
	if err != nil {
		slog.Error("echo start", "err", err)
		panic(err)
	}
}
