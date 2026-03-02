package main

import (
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		slog.Error("Failed to open SQLite Database")
		return
	}

	defer db.Close()
	Migrate(db)
	Insert(db)

	users, err := Query(db)
	if err != nil {
		slog.Error("Failed to get users from DB")
		return
	}

	slog.Info("DB Users: ", users)

	router := NewRouterWithMiddleware()
	MountRoutes(router)
	http.ListenAndServe(":8000", router)
}
