package controllers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sduoduo233/pbb/db"
)

func randomToken() string {
	var buf [16]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

func csrf(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		token := "" // csrf token from cookie

		cookie, err := c.Cookie("CSRF_TOKEN")
		if err != nil || cookie == nil || cookie.Value == "" {
			// csrf token does not exist
			// set a new cookie
			token = randomToken()
			c.SetCookie(&http.Cookie{
				HttpOnly: true,
				Name:     "CSRF_TOKEN",
				Path:     "/",
				Value:    token,
			})
		} else {
			// csrf token exists in cookie
			token = cookie.Value
		}

		if c.Request().Method == http.MethodPost && token != c.FormValue("csrf_token") {
			// request is POST and csrf token is invalid
			return c.HTML(http.StatusBadRequest, "invalid csrf token")
		}

		c.Set("CSRF_TOKEN", token)

		return next(c)
	}
}

func auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("TOKEN")
		if err != nil {
			return next(c)
		}

		var t db.Token
		err = db.DB.Get(&t, "SELECT * FROM tokens WHERE token = $1", cookie.Value)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("db: %w", err)
			}
			return next(c)
		}

		var u db.User
		err = db.DB.Get(&u, "SELECT * FROM users WHERE id = $1", t.UserId)
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}

		c.Set("USER", u)

		return next(c)
	}
}

func mustAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, ok := c.Get("USER").(db.User)
		if !ok {
			c.Response().Header().Set("Refresh", "1; url=/login")
			return c.HTML(http.StatusForbidden, "Authentication required. Redirecting...")
		}

		return next(c)
	}
}

func GetUser(c echo.Context) *db.User {
	u, ok := c.Get("USER").(db.User)
	if !ok {
		return nil
	}

	return &u
}

func MustGetUser(c echo.Context) *db.User {
	u, ok := c.Get("USER").(db.User)
	if !ok {
		panic("user is not in context")
	}

	return &u
}
