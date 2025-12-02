package main

import (
	"log/slog"
	"slices"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sduoduo233/pbb/controllers"
	"github.com/sduoduo233/pbb/cron"
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
				// ignore polling requests
				if !slices.Contains([]string{"/agent/metric", "/agent/services", "/agent/info", "/servers", "/servers/:id/data"}, v.RoutePath) {
					slog.Info("request", "method", v.Method, "uri", v.URI, "status", v.Status)
				}
			} else {
				slog.Info("request", "method", v.Method, "uri", v.URI, "status", v.Status, "error", v.Error)
			}
			return nil
		},
		LogMethod:    true,
		LogURI:       true,
		LogStatus:    true,
		LogError:     true,
		LogRoutePath: true,
	}))

	// template
	e.Renderer = &html.TemplateRenderer{
		Template: html.LoadTemplates(),
	}

	// cron jobs
	cron.Init()

	// controllers
	controllers.Route(e.Group(""))

	err := e.Start(":3005")
	if err != nil {
		slog.Error("echo start", "err", err)
		panic(err)
	}
}
