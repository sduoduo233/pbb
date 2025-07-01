package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
)

func servers(c echo.Context) error {
	servers := make([]db.ServerWithGroupLabel, 0)
	err := db.DB.Select(&servers, "SELECT s.id, s.label, s.group_id, s.hidden, s.last_report, g.label as group_label FROM servers AS s LEFT JOIN groups AS g ON s.group_id = g.id ORDER BY s.id ASC")
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return c.Render(http.StatusOK, "servers", D{"title": "Servers", "servers": servers})
}

func addServer(c echo.Context) error {
	groups := make([]db.Group, 0)
	err := db.DB.Select(&groups, "SELECT * FROM groups ORDER BY id ASC")
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.Render(http.StatusOK, "add_server", D{"title": "Add Server", "groups": groups, "post": D{
		"secret": randomToken(),
	}})
}

func doAddServer(c echo.Context) error {
	label := c.FormValue("label")
	secret := c.FormValue("secret")
	hidden := c.FormValue("hidden") == "yes"
	groupID, err := strconv.Atoi(c.FormValue("group"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid group ID")
	}

	if groupID < 0 {
		_, err = db.DB.Exec("INSERT INTO servers (label, group_id, hidden, secret) VALUES (?, ?, ?, ?)", label, nil, hidden, secret)
	} else {
		_, err = db.DB.Exec("INSERT INTO servers (label, group_id, hidden, secret) VALUES (?, ?, ?, ?)", label, groupID, hidden, secret)
	}
	if err != nil {
		if db.IsUniqueConstraintErr(err) {
			groups := make([]db.Group, 0)
			err := db.DB.Select(&groups, "SELECT * FROM groups ORDER BY id ASC")
			if err != nil {
				return fmt.Errorf("db: %w", err)
			}

			return c.Render(http.StatusOK, "add_server", D{"title": "Add Server", "error": "Server with this label already exists", "post": c.Request().PostForm, "groups": groups})
		}
		return fmt.Errorf("db: %w", err)
	}

	return c.Redirect(http.StatusFound, "/dashboard/servers")
}

func editServer(c echo.Context) error {
	id := c.Param("id")

	server := db.Server{}
	err := db.DB.Get(&server, "SELECT * FROM servers WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Render(http.StatusNotFound, "error", D{"error": "Server Not Found"})
		}
		return fmt.Errorf("db: %w", err)
	}

	groups := make([]db.Group, 0)
	err = db.DB.Select(&groups, "SELECT * FROM groups ORDER BY id ASC")
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.Render(http.StatusOK, "edit_server", D{
		"title": "Edit Server",
		"post": D{
			"label":  server.Label,
			"group":  If(server.GroupId.Valid, server.GroupId.Int32, -1),
			"hidden": If(server.Hidden, "yes", "no"),
			"secret": server.Secret,
		},
		"id":     id,
		"groups": groups,
	})
}

func doEditServer(c echo.Context) error {
	id := c.Param("id")
	label := c.FormValue("label")
	hidden := c.FormValue("hidden") == "yes"
	secret := c.FormValue("secret")
	groupID, err := strconv.Atoi(c.FormValue("group"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid group ID")
	}

	if groupID < 0 {
		_, err = db.DB.Exec("UPDATE servers SET label = ?, group_id = ?, hidden = ?, secret = ? WHERE id = ?", label, nil, hidden, secret, id)
	} else {
		_, err = db.DB.Exec("UPDATE servers SET label = ?, group_id = ?, hidden = ?, secret = ? WHERE id = ?", label, groupID, hidden, secret, id)
	}
	if err != nil {
		if db.IsUniqueConstraintErr(err) {
			groups := make([]db.Group, 0)
			err = db.DB.Select(&groups, "SELECT * FROM groups ORDER BY id ASC")
			if err != nil {
				return fmt.Errorf("db: %w", err)
			}
			return c.Render(http.StatusOK, "edit_server", D{"title": "Edit Server", "error": "Server with this label already exists", "post": c.Request().PostForm, "id": id, "groups": groups})
		}
		return fmt.Errorf("db: %w", err)
	}

	return c.Redirect(http.StatusFound, "/dashboard/servers")
}

func deleteServer(c echo.Context) error {
	id := c.Param("id")

	_, err := db.DB.Exec("DELETE FROM servers WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Render(http.StatusNotFound, "error", D{"error": "Server Not Found"})
		}
		return fmt.Errorf("db: %w", err)
	}

	return c.String(http.StatusOK, "OK")
}

func installAgent(c echo.Context) error {
	id := c.Param("id")

	var secret string
	err := db.DB.Get(&secret, "SELECT secret FROM servers WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Render(http.StatusNotFound, "error", D{"error": "Server Not Found"})
		}
		return fmt.Errorf("db: %w", err)
	}

	var publicUrl string
	err = db.DB.Get(&publicUrl, "SELECT value FROM settings WHERE key = 'public_url'")
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.Render(http.StatusOK, "install_agent", D{"title": "Install Agent", "secret": secret, "public_url": publicUrl})
}
