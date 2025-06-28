package db

import (
	"log/slog"

	_ "embed"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
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
		_, err = DB.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", "admin@example.com", string(hashed))
		if err != nil {
			panic(err)
		}
		slog.Warn("admin user created", "email", "admin@example.com", "password", "admin")
	}
}

func IsUniqueConstraintErr(err error) bool {
	if err == nil {
		return false
	}
	if sqliteErr, ok := err.(sqlite3.Error); ok {
		return sqliteErr.ExtendedCode == 2067 // SQLITE_CONSTRAINT_UNIQUE
	}
	return false
}
