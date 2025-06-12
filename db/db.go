package db

import (
	"encoding/hex"
	"log/slog"

	_ "embed"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var DB *sqlx.DB

//go:embed init.sql
var initSql string

func Init() {
	var err error
	DB, err = sqlx.Connect("sqlite3", "./sqlite.db")
	if err != nil {
		slog.Error("connect database", "err", err)
		panic(err)
	}

	_, err = DB.Exec(initSql)
	if err != nil {
		slog.Error("create tables", "err", err)
		panic(err)
	}

	// create default user
	var cnt int
	err = DB.Get(&cnt, "SELECT COUNT(*) FROM users")
	if err != nil {
		panic(err)
	}

	if cnt == 0 {
		hashed, err := bcrypt.GenerateFromPassword([]byte("admin"), 0)
		if err != nil {
			panic(err)
		}
		_, err = DB.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", "admin@example.com", hex.EncodeToString(hashed))
		if err != nil {
			panic(err)
		}
		slog.Warn("admin user created", "email", "admin@example.com", "password", "admin")
	}
}
