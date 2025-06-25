package cron

import (
	"log/slog"
	"time"

	"github.com/sduoduo233/pbb/db"
)

func sample() {
	var m []db.ServerMetrics
	err := db.DB.Select(
		&m,
		`SELECT
			server_id,
			AVG(cpu) as cpu,
			AVG(memory_percent) as memory_percent,
			CAST(AVG(memory_used) AS INTEGER) as memory_used,
			CAST(AVG(memory_total) AS INTEGER) as memory_total,
			AVG(disk_percent) as disk_percent,
			CAST(AVG(disk_used) AS INTEGER) as disk_used,
			CAST(AVG(disk_total) AS INTEGER) as disk_total,
			CAST(AVG(network_out_rate) AS INTEGER) as network_out_rate,
			CAST(AVG(network_in_rate) AS INTEGER) as network_in_rate,
			AVG(swap_percent) as swap_percent,
			CAST(AVG(swap_used) AS INTEGER) as swap_used,
			CAST(AVG(swap_total) AS INTEGER) as swap_total,
			MAX(uptime) as uptime,
			0.0 as load_1,
			0.0 as load_5,
			0.0 as load_15
		FROM server_metrics WHERE created_at > ? GROUP BY server_id`,
		time.Now().Add(-10*time.Minute).Unix(),
	)
	if err != nil {
		slog.Error("sample server_metrics", "err", err)
		return
	}

	for _, v := range m {
		v.CreatedAt = uint64(time.Now().Unix())
		_, err := db.DB.NamedExec(`INSERT INTO server_metrics_10m (created_at, server_id, cpu, memory_percent, memory_used, memory_total, disk_percent, disk_used, disk_total, network_out_rate, network_in_rate, swap_percent, swap_used, swap_total, uptime, load_1, load_5, load_15) VALUES (:created_at, :server_id, :cpu, :memory_percent, :memory_used, :memory_total, :disk_percent, :disk_used, :disk_total, :network_out_rate, :network_in_rate, :swap_percent, :swap_used, :swap_total, :uptime, :load_1, :load_5, :load_15)`, v)
		if err != nil {
			slog.Error("insert server_metrics_10m", "err", err, "server_id", v.ServerId)
			continue
		}
	}

	slog.Info("sampled server metrics")
}
