package cron

import (
	"log/slog"
	"time"

	"github.com/sduoduo233/pbb/db"
)

func clean() {

	r, err := db.DB.Exec("DELETE FROM server_metrics WHERE created_at < ?", time.Now().Add(-24*7*time.Hour).Unix()) // delete metrics older than 7 days
	if err != nil {
		slog.Error("clean server_metrics", "err", err)
		return
	}

	rowsAffected, err := r.RowsAffected()
	if err != nil {
		slog.Error("clean server_metrics", "err", err)
		return
	}

	slog.Info("old records deleted", "rows deleted", rowsAffected)
}
