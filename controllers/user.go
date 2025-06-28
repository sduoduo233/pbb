package controllers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
	"golang.org/x/crypto/bcrypt"
)

func user(c echo.Context) error {
	return c.Render(http.StatusOK, "user", D{"title": "User", "email": MustGetUser(c).Email})
}

func doChangePassword(c echo.Context) error {
	u := MustGetUser(c)

	email := c.FormValue("email")
	current := c.FormValue("current")
	newPassword := c.FormValue("password1")
	confirmPassword := c.FormValue("password2")

	if current != "" || newPassword != "" || confirmPassword != "" {

		if newPassword != confirmPassword {
			return c.Render(http.StatusBadRequest, "user", D{
				"title": "User",
				"error": "New password and confirmation do not match.",
			})
		}
		if len(newPassword) < 8 {
			return c.Render(http.StatusBadRequest, "user", D{
				"title": "User",
				"error": "New password must be at least 8 characters long.",
			})
		}

		var currentHash string
		err := db.DB.Get(&currentHash, "SELECT password FROM users WHERE id = ?", u.Id)
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}

		err = bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(current))
		if err != nil {
			return c.Render(http.StatusBadRequest, "user", D{
				"title": "User",
				"error": "Current password is incorrect.",
			})
		}

		newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("bcrypt: %w", err)
		}

		_, err = db.DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(newHash), u.Id)
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}

	}

	_, err := db.DB.Exec("UPDATE users SET email = ? WHERE id = ?", email, u.Id)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	return c.Render(http.StatusOK, "user", D{
		"title":   "User",
		"success": "User profile updated.",
	})
}
