package controllers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
)

func settings(c echo.Context) error {
	var settings []db.Setting
	err := db.DB.Select(&settings, "SELECT * FROM settings")
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	settingsMap := make(map[string]string, len(settings))
	for _, setting := range settings {
		settingsMap[setting.Key] = setting.Value
	}

	return c.Render(http.StatusOK, "settings", D{"settings": settingsMap, "title": "Settings"})
}

func doUpdateSettings(c echo.Context) error {

	c.Request().ParseForm()

	for key, values := range c.Request().PostForm {
		if len(values) == 0 {
			continue
		}

		value := values[0]
		_, err := db.DB.Exec("UPDATE settings SET value = ? WHERE key = ?", value, key)
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}
	}

	return c.Redirect(http.StatusSeeOther, "/dashboard/settings")
}
