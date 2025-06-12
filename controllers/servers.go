package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func servers(c echo.Context) error {
	return c.Render(http.StatusOK, "servers", D{"title": "Servers"})
}

func addServer(c echo.Context) error {
	return c.Render(http.StatusOK, "add_server", D{"title": "Add Server"})
}
