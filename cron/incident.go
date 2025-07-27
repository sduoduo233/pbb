package cron

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/sduoduo233/pbb/controllers/types"
	"github.com/sduoduo233/pbb/db"
)

func incident() {
	type isOnline struct {
		Id     int32 `db:"id"`
		Online bool  `db:"online"`
	}
	var online []isOnline
	err := db.DB.Select(&online, "SELECT id, last_report > ? AS online FROM servers WHERE last_report IS NOT NULL", time.Now().Add(-time.Second*30).Unix())
	if err != nil {
		slog.Error("incident check", "err", fmt.Errorf("db: %w", err))
		return
	}

	for _, v := range online {
		var incident db.Incident
		err := db.DB.Get(&incident, "SELECT * FROM incidents WHERE server_id = ? ORDER BY id DESC LIMIT 1", v.Id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			slog.Error("incident check", "err", fmt.Errorf("db: %w", err), "server_id", v.Id)
			continue
		}

		if (errors.Is(err, sql.ErrNoRows) && !v.Online) || (incident.State == db.IncidentStateResolved && !v.Online) {
			// no incident found, but server is offline -> create new incident
			// latest incident was resolved but server is online -> create new incident

			incident.ServerId = v.Id
			incident.StartedAt = time.Now().Unix()
			incident.EndedAt = types.NullInt64{NullInt64: sql.NullInt64{Valid: false}}
			incident.State = db.IncidentStateOngoing
			_, err = db.DB.NamedExec(`INSERT INTO incidents (server_id, started_at, ended_at, state) VALUES (:server_id, :started_at, :ended_at, :state)`, &incident)
			if err != nil {
				slog.Error("incident check", "err", fmt.Errorf("db insert: %w", err), "server_id", v.Id)
				continue
			}
			slog.Info("incident created", "server_id", v.Id, "state", incident.State, "started_at", incident.StartedAt)
			continue
		}
		if incident.State == db.IncidentStateOngoing && v.Online {
			// latest incident is ongoing but server is online -> mark as resolved
			_, err = db.DB.Exec("UPDATE incidents SET state = ?, ended_at = ? WHERE id = ?", db.IncidentStateResolved, time.Now().Unix(), incident.Id)
			if err != nil {
				slog.Error("incident check", "err", fmt.Errorf("db update: %w", err), "server_id", v.Id)
				continue
			}
			slog.Info("incident resolved", "server_id", v.Id, "state", db.IncidentStateResolved, "ended_at", time.Now().Unix())
			continue
		}
	}
}
