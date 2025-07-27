package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
)

func network(c echo.Context) error {
	id := c.Param("id")

	user := GetUser(c)
	showHidden := user != nil

	var server db.Server
	err := db.DB.Get(&server, "SELECT * FROM servers WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Render(http.StatusNotFound, "error", D{"error": "404 Not Found"})
		}
		return fmt.Errorf("db: %w", err)
	}

	if !showHidden && server.Hidden {
		return c.Render(http.StatusNotFound, "error", D{"error": "404 Not Found"})
	}

	var relatedServices = make([]db.Service, 0)
	err = db.DB.Select(&relatedServices, "SELECT s.* FROM services s INNER JOIN server_services ss ON s.id = ss.service_id AND ss.server_id = ?", id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	duration := c.QueryParam("duration")
	if duration == "" {
		duration = "3h"
	}
	start := time.Now().Add(-3 * time.Hour).Unix()
	switch c.QueryParam("duration") {
	case "3h":
		start = time.Now().Add(-3 * time.Hour).Unix()
	case "30h":
		start = time.Now().Add(-30 * time.Hour).Unix()
	case "3d":
		start = time.Now().Add(-3 * 24 * time.Hour).Unix()
	}

	return c.Render(http.StatusOK, "network", D{"id": id, "services": relatedServices, "end": time.Now().Unix(), "start": start, "duration": duration})
}
