package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
)

func view(c echo.Context) error {
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

	return c.Render(http.StatusOK, "view", D{"label": server.Label, "id": id})
}

func viewData(c echo.Context) error {
	id := c.Param("id")

	duration, err := strconv.Atoi(c.QueryParam("duration"))
	if err != nil || duration < 60*5 || duration > 60*60*24*3 {
		duration = 60 * 5
	}

	user := GetUser(c)
	showHidden := user != nil

	var server db.Server
	err = db.DB.Get(&server, "SELECT * FROM servers WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Render(http.StatusNotFound, "error", D{"error": "404 Not Found"})
		}
		return fmt.Errorf("db: %w", err)
	}

	if !showHidden && server.Hidden {
		return c.Render(http.StatusNotFound, "error", D{"error": "404 Not Found"})
	}

	metrics := make([]db.ServerMetrics, 0)
	after := time.Now().Add(-(time.Second * time.Duration(duration))).Unix()
	err = db.DB.Select(&metrics, "SELECT * FROM server_metrics WHERE server_id = ? AND created_at > ? ORDER BY id DESC", server.Id, after)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.JSON(http.StatusOK, D{"server": D{"last_report": server.LastReport.Int64, "arch": server.Arch.String, "operating_system": server.OS.String, "cpu": server.Cpu.String}, "metrics": metrics})
}
