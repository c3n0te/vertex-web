package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file: ", "error", err)
		return
	}

	key := os.Getenv("SERVER_SECRET_KEY")
	if key == "" {
		slog.Error("Error loading Server Key env var.")
		return
	}

	db, err := sqlx.Connect("sqlite3", "./vertex.db")
	if err != nil {
		slog.Error("Failed to open SQLite Database: ", "error", err)
		return
	}

	defer db.Close()
	srv := NewServer(key, db)
	err = srv.InitDB()
	if err != nil {
		slog.Error("Failed to initialize SQLite Database: ", "error", err)
		return
	}

	srv.ServeAssets()
	srv.MountHandlers()
	http.ListenAndServe(":8000", srv.Router)
}
