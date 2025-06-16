package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func view(c echo.Context) error {
	return c.Render(http.StatusOK, "view", D{})
}
