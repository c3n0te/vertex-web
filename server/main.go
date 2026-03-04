package main

import (
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file")
	}

	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		slog.Error("Failed to open SQLite Database")
		return
	}

	defer db.Close()
	Migrate(db)
	InsertSats(db)

	router := NewRouter()
	MountRoutes(router, db)
	http.ListenAndServe(":8000", router)
}
