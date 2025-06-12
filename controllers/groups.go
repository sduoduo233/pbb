package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
)

func groups(c echo.Context) error {
	groups := make([]db.Group, 0)
	err := db.DB.Select(&groups, "SELECT * FROM groups ORDER BY id ASC")
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	return c.Render(http.StatusOK, "groups", D{"title": "Groups", "groups": groups})
}

func addGroup(c echo.Context) error {
	return c.Render(http.StatusOK, "add_group", D{"title": "Add Group"})
}

func doAddGroup(c echo.Context) error {
	label := c.FormValue("label")
	hidden := c.FormValue("hidden") == "on"

	_, err := db.DB.Exec("INSERT INTO groups (label, hidden) VALUES (?, ?)", label, hidden)
	if err != nil {
		if db.IsUniqueConstraintErr(err) {
			return c.Render(http.StatusOK, "add_group", D{"title": "Add Group", "error": "Group with this label already exists", "post": c.Request().PostForm})
		}
		return fmt.Errorf("db: %w", err)
	}

	return c.Redirect(http.StatusFound, "/dashboard/groups")
}

func editGroup(c echo.Context) error {
	id := c.Param("id")
	group := db.Group{}

	err := db.DB.Get(&group, "SELECT * FROM groups WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Render(http.StatusNotFound, "error", D{"error": "Group Not Found"})
		}
		return fmt.Errorf("db: %w", err)
	}

	return c.Render(http.StatusOK, "edit_group", D{"title": "Edit Group", "post": D{"label": group.Label, "hidden": group.Hidden}, "id": id})
}

func doEditGroup(c echo.Context) error {
	id := c.Param("id")
	label := c.FormValue("label")
	hidden := c.FormValue("hidden") == "on"

	_, err := db.DB.Exec("UPDATE groups SET label = ?, hidden = ? WHERE id = ?", label, hidden, id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.Redirect(http.StatusFound, "/dashboard/groups")
}

func deleteGroup(c echo.Context) error {
	id := c.Param("id")

	_, err := db.DB.Exec("DELETE FROM groups WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.String(http.StatusOK, "OK")
}
