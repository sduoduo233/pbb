package controllers

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
)

func viewIncidents(c echo.Context) error {
	user := GetUser(c)
	ids := c.QueryParams()["id"]

	incidents := make([]db.IncidentWithServerLabel, 0, len(ids))

	for _, id := range ids {
		var incident db.IncidentWithServerLabel
		err := db.DB.Get(&incident, "SELECT incidents.id, servers.label AS server_label, incidents.server_id, incidents.started_at, incidents.ended_at, incidents.state, servers.hidden AS hidden FROM incidents LEFT JOIN servers ON incidents.server_id = servers.id WHERE incidents.id = ?", id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return fmt.Errorf("db: %w", err)
		}
		if !incident.Hidden || user != nil {
			incidents = append(incidents, incident)
		}
	}

	return c.Render(200, "incidents", D{
		"incidents": incidents,
		"title":     "Incidents",
	})
}
