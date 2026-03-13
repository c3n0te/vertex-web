package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var tokenAuth *jwtauth.JWTAuth

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

	tokenAuth = jwtauth.New("HS256", []byte(key), nil)
	db, err := sqlx.Connect("sqlite3", "./vertex.db")
	if err != nil {
		slog.Error("Failed to open SQLite Database: ", "error", err)
		return
	}

	defer db.Close()
	srv := NewServer(key, db, tokenAuth)
	err = srv.InitDB()
	router := NewRouter()
	ServeAssets(router)

	router.Get("/", srv.GetHome)
	router.Get("/auth/login", srv.GetAuthLogin)
	router.Get("/auth/sign-up", srv.GetSignUp)
	router.Post("/auth/sign-up", srv.PostSignUp)
	router.Post("/auth/login", srv.PostLogin)

	router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(srv.TokenAuth))
		r.Use(jwtauth.Authenticator(srv.TokenAuth))
		r.Get("/dashboard", srv.GetDashboard)
		r.Get("/plan", srv.GetPlan)
		r.Post("/plan/submit", srv.PostPlanSubmit)
		r.Get("/schedule", srv.GetSchedule)
	})

	http.ListenAndServe(":8000", router)
}
