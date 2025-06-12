package controllers

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
	"golang.org/x/crypto/bcrypt"
)

func login(c echo.Context) error {
	user := GetUser(c)
	if user != nil {
		return c.Redirect(http.StatusFound, "/dashboard/servers")
	}

	return c.Render(http.StatusOK, "login", D{"title": "Login"})
}

func doLogin(c echo.Context) error {
	user := GetUser(c)
	if user != nil {
		return c.Redirect(http.StatusFound, "/dashboard/servers")
	}

	email := strings.ToLower(c.Request().PostFormValue("email"))
	password := c.Request().PostFormValue("password")

	var u db.User
	err := db.DB.Get(&u, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Render(http.StatusOK, "login", D{"title": "Login", "error": "Wrong email or password"})
		}
		return fmt.Errorf("db: %w", err)
	}

	// compare password

	hashed, err := hex.DecodeString(u.Password)
	if err != nil {
		return fmt.Errorf("bad password hash: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(hashed, []byte(password))
	if err != nil {
		return c.Render(http.StatusOK, "login", D{"title": "Login", "error": "Wrong email or password"})
	}

	// issue token

	token := randomToken()
	_, err = db.DB.Exec("INSERT INTO tokens (token, user_id, created_at) VALUES ($1, $2, $3)", token, u.Id, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	c.SetCookie(&http.Cookie{
		Name:     "TOKEN",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})

	return c.Redirect(http.StatusFound, "/dashboard/servers")
}
