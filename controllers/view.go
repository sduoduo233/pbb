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
	if err != nil || duration < 60*5 || duration > 60*60*24*7 {
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

	if duration > 7200 { // 2 hours more
		// use sampled data
		err = db.DB.Select(&metrics, "SELECT * FROM server_metrics_10m WHERE server_id = ? AND created_at > ? ORDER BY id DESC", server.Id, after)
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}
	} else {
		err = db.DB.Select(&metrics, "SELECT * FROM server_metrics WHERE server_id = ? AND created_at > ? ORDER BY id DESC", server.Id, after)
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}
	}

	var latest db.ServerMetrics
	err = db.DB.Get(&latest, "SELECT * FROM server_metrics WHERE server_id = ? AND created_at > ? ORDER BY id DESC LIMIT 1", server.Id, time.Now().Add(-time.Second*10).Unix())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("db: %w", err)
	}

	incidents := make([]db.Incident, 0)
	err = db.DB.Select(&incidents, "SELECT * FROM incidents WHERE server_id = ? AND ((ended_at IS NOT NULL AND ended_at > ?) OR (started_at > ?) OR state = 'ongoing') ORDER BY id ASC", server.Id, after, after)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.JSON(http.StatusOK, D{
		"server": D{
			"last_report":      server.LastReport.Int64,
			"arch":             server.Arch.String,
			"operating_system": server.OS.String,
			"cpu":              server.Cpu.String,
			"online":           server.LastReport.Valid && time.Since(time.Unix(server.LastReport.Int64, 0)) < time.Second*10,
		},
		"metrics":   metrics,
		"latest":    latest,
		"incidents": incidents,
	})
}
