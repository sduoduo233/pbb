package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
)

func services(c echo.Context) error {

	var ss []db.Service

	err := db.DB.Select(&ss, "SELECT * FROM services")
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.Render(http.StatusOK, "services", D{"title": "Services", "services": ss})
}

func addService(c echo.Context) error {

	return c.Render(http.StatusOK, "add_service", D{"title": "Add Service", "post": D{"type": "ping"}})
}

func doAddService(c echo.Context) error {

	label := c.FormValue("label")
	serviceType := c.FormValue("type")
	host := c.FormValue("host")

	result, err := db.DB.Exec("INSERT INTO services (label, type, host) VALUES (?, ?, ?)", label, serviceType, host)
	if err != nil {
		if db.IsUniqueConstraintErr(err) {
			return c.Render(http.StatusOK, "add_service", D{"title": "Add Service", "error": "Service with this label already exists"})
		}
		return fmt.Errorf("db: %w", err)
	}

	lastId, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.Redirect(http.StatusFound, "/dashboard/services/"+strconv.Itoa(int(lastId)))
}

func deleteService(c echo.Context) error {
	id := c.Param("id")

	_, err := db.DB.Exec("DELETE FROM services WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.String(http.StatusNotFound, "server not found")
		}
		return fmt.Errorf("db: %w", err)
	}

	return c.String(http.StatusOK, "OK")
}

func editService(c echo.Context) error {
	type ServerSelected struct {
		Id          int32  `db:"id"`
		ServerLabel string `db:"server_label"`
		Selected    bool   `db:"selected"`
	}

	id := c.Param("id")

	var s db.Service
	var ss []ServerSelected

	err := db.DB.Get(&s, "SELECT * FROM services WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Render(http.StatusNotFound, "error", D{"error": "Service Not Found"})
		}
		return fmt.Errorf("db: %w", err)
	}

	err = db.DB.Select(&ss, "SELECT s.id AS id, s.label AS server_label, ss.id IS NOT NULL AS selected FROM servers s LEFT JOIN server_services ss ON s.id = ss.server_id AND ss.service_id = ?", id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.Render(http.StatusOK, "edit_service", D{
		"title": "Edit Service",
		"post": D{
			"label": s.Label,
			"host":  s.Host,
			"type":  s.Type,
		},
		"id":      id,
		"servers": ss,
	})

}

func doEditService(c echo.Context) error {
	id := c.Param("id")

	label := c.FormValue("label")
	serviceType := c.FormValue("type")
	host := c.FormValue("host")

	_, err := db.DB.Exec("UPDATE services SET label = ?, type = ?, host = ? WHERE id = ?", label, serviceType, host, id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	_, err = db.DB.Exec("DELETE FROM server_services WHERE service_id = ?", id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	for k, v := range c.Request().PostForm {
		if !strings.HasPrefix(k, "server-") || len(v) != 1 || v[0] != "on" {
			continue
		}
		serverId, err := strconv.Atoi(strings.TrimPrefix(k, "server-"))
		if err != nil {
			continue
		}
		_, err = db.DB.Exec("INSERT INTO server_services (server_id, service_id) VALUES (?, ?)", serverId, id)
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}
	}

	return c.Redirect(http.StatusFound, "/dashboard/services")
}
