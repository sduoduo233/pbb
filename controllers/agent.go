package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
)

func metrics(c echo.Context) error {
	secret := c.Request().Header.Get("x-secret")

	var s db.Server
	err := db.DB.Get(&s, "SELECT * FROM servers WHERE secret = ?", secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.String(http.StatusUnauthorized, "bad x-secret")
		}
		return fmt.Errorf("get server %w", err)
	}

	slog.Debug("new metric", "server", s.Id, "label", s.Label)

	var m db.ServerMetrics
	err = c.Bind(&m)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request: "+err.Error())
	}

	m.ServerId = s.Id
	m.CreatedAt = uint64(time.Now().Unix())

	_, err = db.DB.NamedExec("INSERT INTO server_metrics (created_at, server_id, cpu, memory_percent, memory_used, memory_total, disk_percent, disk_used, disk_total, network_out_rate, network_in_rate, swap_used, swap_total, swap_percent) VALUES (:created_at, :server_id, :cpu, :memory_percent, :memory_used, :memory_total, :disk_percent, :disk_used, :disk_total, :network_out_rate, :network_in_rate, :swap_used, :swap_total, :swap_percent)", &m)
	if err != nil {
		return fmt.Errorf("sql 1: %w", err)
	}

	_, err = db.DB.Exec("UPDATE servers SET last_report = ? WHERE id = ?", time.Now().Unix(), s.Id)
	if err != nil {
		return fmt.Errorf("sql 2: %w", err)
	}

	return c.String(http.StatusOK, "OK")
}
