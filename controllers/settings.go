package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
	"github.com/sduoduo233/pbb/telegram"
)

func settings(c echo.Context) error {
	tab := c.QueryParam("tab")
	if tab != "telegram" {
		tab = "general"
	}
	return renderSettings(c, tab, D{})
}

func renderSettings(c echo.Context, tab string, data D) error {
	var settings []db.Setting
	err := db.DB.Select(&settings, "SELECT * FROM settings")
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	settingsMap := make(map[string]string, len(settings))
	for _, setting := range settings {
		settingsMap[setting.Key] = setting.Value
	}

	data["settings"] = settingsMap
	data["title"] = "Settings"
	data["tab"] = tab

	if tab == "telegram" {
		telegramUsers := make([]db.TelegramUser, 0)
		err = db.DB.Select(&telegramUsers, "SELECT * FROM telegram_users ORDER BY id ASC")
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}
		data["telegramUsers"] = telegramUsers
	}

	return c.Render(http.StatusOK, "settings", data)
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

func doUpdateTelegramSettings(c echo.Context) error {
	token := strings.TrimSpace(c.FormValue("telegram_bot_token"))
	notify := If(c.FormValue("telegram_notify") == "on", "1", "0")

	_, err := db.DB.Exec("UPDATE settings SET value = ? WHERE key = 'telegram_bot_token'", token)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	_, err = db.DB.Exec("UPDATE settings SET value = ? WHERE key = 'telegram_notify'", notify)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	telegram.Reload()

	return c.Redirect(http.StatusSeeOther, "/dashboard/settings?tab=telegram")
}

func doAddTelegramUser(c echo.Context) error {
	userId, err := strconv.ParseInt(strings.TrimSpace(c.FormValue("user_id")), 10, 64)
	if err != nil {
		return renderSettings(c, "telegram", D{"error": "Invalid user ID"})
	}

	_, err = db.DB.Exec("INSERT INTO telegram_users (user_id) VALUES (?)", userId)
	if err != nil {
		if db.IsUniqueConstraintErr(err) {
			return renderSettings(c, "telegram", D{"error": "User already added"})
		}
		return fmt.Errorf("db: %w", err)
	}

	return c.Redirect(http.StatusSeeOther, "/dashboard/settings?tab=telegram")
}

func deleteTelegramUser(c echo.Context) error {
	id := c.Param("id")

	_, err := db.DB.Exec("DELETE FROM telegram_users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.String(http.StatusOK, "OK")
}
