package main

import (
	"log/slog"
	"net/http"
	"vertex/ui/pages"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file: ", "error", err)
	}

	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		slog.Error("Failed to open SQLite Database: ", "error", err)
		return
	}

	defer db.Close()
	Migrate(db)
	InsertStns(db)
	InsertSats(db)

	router := NewRouter()
	ServeAssets(router)
	home := pages.HomePage()
	login := pages.LoginPage()
	router.Handle("/", templ.Handler(home))
	router.Handle("/auth/login", templ.Handler(login))

	router.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		stns, err := ReadStns(db)
		if err != nil {
			slog.Error("Failed to read stations: ", "error", err)
		}

		sats, err := ReadSats(db)
		if err != nil {
			slog.Error("Failed to read satellites: ", "error", err)
		}

		dashboard := pages.DashboardPage(stns, sats)
		templ.Handler(dashboard).ServeHTTP(w, r)
	})

	router.Get("/plan", func(w http.ResponseWriter, r *http.Request) {
		sats, err := ReadSats(db)
		if err != nil {
			slog.Error("Failed to read satellites: ", "error", err)
		}

		plan := pages.PlanPage(sats)
		templ.Handler(plan).ServeHTTP(w, r)
	})

	router.Get("/schedule", func(w http.ResponseWriter, r *http.Request) {
		schedule := pages.SchedulePage()
		templ.Handler(schedule).ServeHTTP(w, r)
	})

	router.Post("/plan/submit", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			slog.Error("Failed to parse plan form: ", "error", err)
		}

		InsertTask(db, r.Form)
		http.Redirect(w, r, "/plan", http.StatusSeeOther)
	})

	router.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)
	})

	http.ListenAndServe(":8000", router)
}
