package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/controllers/types"
	"github.com/sduoduo233/pbb/db"
)

func findServerBySecret(secret string) (*db.Server, error) {
	var s db.Server
	err := db.DB.Get(&s, "SELECT * FROM servers WHERE secret = ?", secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find server: db: %w", err)
	}

	return &s, nil
}

func metrics(c echo.Context) error {
	secret := c.Request().Header.Get("x-secret")
	s, err := findServerBySecret(secret)
	if err != nil {
		return err
	}
	if s == nil {
		return c.String(http.StatusUnauthorized, "bad x-secret")
	}

	slog.Debug("new metric", "server", s.Id, "label", s.Label)

	var m db.ServerMetrics
	err = c.Bind(&m)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request: "+err.Error())
	}

	m.ServerId = s.Id
	m.CreatedAt = uint64(time.Now().Unix())

	_, err = db.DB.NamedExec("INSERT INTO server_metrics (created_at, server_id, cpu, memory_percent, memory_used, memory_total, disk_percent, disk_used, disk_total, network_out_rate, network_in_rate, swap_used, swap_total, swap_percent, uptime, load_1, load_5, load_15) VALUES (:created_at, :server_id, :cpu, :memory_percent, :memory_used, :memory_total, :disk_percent, :disk_used, :disk_total, :network_out_rate, :network_in_rate, :swap_used, :swap_total, :swap_percent, :uptime, :load_1, :load_5, :load_15)", &m)
	if err != nil {
		return fmt.Errorf("db 1: %w", err)
	}

	_, err = db.DB.Exec("UPDATE servers SET last_report = ? WHERE id = ?", time.Now().Unix(), s.Id)
	if err != nil {
		return fmt.Errorf("db 2: %w", err)
	}

	return c.String(http.StatusOK, "OK")
}

func systemInfo(c echo.Context) error {
	secret := c.Request().Header.Get("x-secret")
	s, err := findServerBySecret(secret)
	if err != nil {
		return err
	}
	if s == nil {
		return c.String(http.StatusUnauthorized, "bad x-secret")
	}

	var i types.ServerInfo
	err = c.Bind(&i)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request: "+err.Error())
	}

	_, err = db.DB.Exec("UPDATE servers SET arch = ?, operating_system = ?, cpu = ?, version = ? WHERE id = ?", i.Arch, i.OS, i.Cpu, i.Version, s.Id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.String(http.StatusOK, "OK")

}

func service(c echo.Context) error {
	secret := c.Request().Header.Get("x-secret")
	s, err := findServerBySecret(secret)
	if err != nil {
		return err
	}
	if s == nil {
		return c.String(http.StatusUnauthorized, "bad x-secret")
	}

	var metrics []types.ServiceMetric
	err = c.Bind(&metrics)
	if err != nil {
		return c.String(http.StatusBadRequest, "bad request: "+err.Error())
	}

	for _, m := range metrics {
		dbMetric := db.ServiceMetric{
			CreatedAt: uint64(time.Now().Unix()),
			Timestamp: m.Timestamp,
			From:      s.Id,
			To:        m.To,
			Min:       m.Min,
			Max:       m.Max,
			Avg:       m.Avg,
			Median:    m.Median,
			Loss:      m.Loss,
		}
		_, err = db.DB.NamedExec("INSERT INTO service_metrics (created_at, timestamp, `from`, `to`, min, max, loss, avg, median) VALUES (:created_at, :timestamp, :from, :to, :min, :max, :loss, :avg, :median)", dbMetric)
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}
	}

	var relatedServices = make([]db.Service, 0)
	err = db.DB.Select(&relatedServices, "SELECT s.* FROM services s INNER JOIN server_services ss ON s.id = ss.service_id AND ss.server_id = ?", s.Id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.JSON(http.StatusOK, relatedServices)
}
