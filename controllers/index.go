package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
)

func indexServers(c echo.Context) error {
	user := GetUser(c)
	showHidden := user != nil

	view := c.QueryParam("view")
	if view != "cards" && view != "table" {
		view = "cards"
	}

	type Server struct {
		ID     int32
		Label  string
		Online bool

		Cpu                 float32
		MemoryUsed          uint64
		MemoryUsedFormated  string
		MemoryTotal         uint64
		MemoryTotalFormated string
		DiskUsed            uint64
		DiskUsedFormated    string
		DiskTotal           uint64
		DiskTotalFormated   string
		NetworkInFormated   string
		NetworkOutFormated  string
		SwapTotal           uint64
		SwapUsed            uint64
		SwapTotalFormated   string
		SwapUsedFormated    string
	}

	type ServerGroup struct {
		ID      int32
		Label   string
		Servers []Server
	}

	statOnline := 0
	statOffline := 0
	statTotal := 0
	statNetworkIn := 0
	statNetworkOut := 0

	groups := make([]ServerGroup, 0)

	var dbGroups []db.Group
	err := db.DB.Select(&dbGroups, "SELECT * FROM groups WHERE (NOT hidden OR ?)", showHidden)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	dbGroups = slices.Insert(dbGroups, 0, db.Group{Id: -1, Label: "Default group"})

	for _, dbGroup := range dbGroups {
		group := ServerGroup{
			ID:      dbGroup.Id,
			Label:   dbGroup.Label,
			Servers: make([]Server, 0),
		}

		var dbServers []db.Server
		err = db.DB.Select(&dbServers, "SELECT * FROM servers WHERE (NOT hidden OR ?) AND (group_id = ? OR (? < 0 AND group_id IS NULL))", showHidden, dbGroup.Id, dbGroup.Id)
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}

		for _, dbServer := range dbServers {
			server := Server{
				ID:     dbServer.Id,
				Label:  dbServer.Label,
				Online: dbServer.LastReport.Valid && time.Since(time.Unix(dbServer.LastReport.Int64, 0)) < time.Second*10,
			}

			if !server.Online {
				statOffline++
			} else {
				statOnline++
			}
			statTotal++

			var metric db.ServerMetrics
			err = db.DB.Get(&metric, "SELECT * FROM server_metrics WHERE server_id = ? ORDER BY id DESC LIMIT 1", dbServer.Id)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					server.MemoryUsedFormated = "0"
					server.MemoryTotalFormated = "0"
					server.DiskUsedFormated = "0"
					server.DiskTotalFormated = "0"
					server.NetworkInFormated = "0 B"
					server.NetworkOutFormated = "0 B"
					group.Servers = append(group.Servers, server)
					continue
				}
				return fmt.Errorf("db: %w", err)
			}

			server.Cpu = metric.Cpu
			server.MemoryUsed = metric.MemoryUsed
			server.MemoryUsedFormated = formatBytes(metric.MemoryUsed)
			server.MemoryTotal = metric.MemoryTotal
			server.MemoryTotalFormated = formatBytes(metric.MemoryTotal)
			server.DiskUsed = metric.DiskUsed
			server.DiskUsedFormated = formatBytes(metric.DiskUsed)
			server.DiskTotal = metric.DiskTotal
			server.DiskTotalFormated = formatBytes(metric.DiskTotal)
			server.NetworkInFormated = formatBytes(metric.NetworkInRate)
			server.NetworkOutFormated = formatBytes(metric.NetworkOutRate)
			server.SwapUsed = metric.SwapUsed
			server.SwapTotal = metric.SwapTotal
			server.SwapTotalFormated = formatBytes(metric.SwapTotal)
			server.SwapUsedFormated = formatBytes(metric.SwapUsed)

			statNetworkIn += int(metric.NetworkInRate)
			statNetworkOut += int(metric.NetworkOutRate)

			group.Servers = append(group.Servers, server)
		}

		groups = append(groups, group)
	}

	if view == "table" {
		return c.Render(http.StatusOK, "index_servers_table", D{"groups": groups, "stat_online": statOnline, "stat_offline": statOffline, "stat_total": statTotal, "stat_network_in": formatBytes(uint64(statNetworkIn)), "stat_network_out": formatBytes(uint64(statNetworkOut))})
	}
	return c.Render(http.StatusOK, "index_servers", D{"groups": groups, "stat_online": statOnline, "stat_offline": statOffline, "stat_total": statTotal, "stat_network_in": formatBytes(uint64(statNetworkIn)), "stat_network_out": formatBytes(uint64(statNetworkOut))})
}

func index(c echo.Context) error {

	view := c.QueryParam("view")
	if view != "cards" && view != "table" {
		view = "cards"
	}

	return c.Render(http.StatusOK, "index", D{"view": view})
}
