package main

import (
	"net/http"

	"vertex/ui/pages"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "valid-token" {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func MountRoutes(router chi.Router, db *sqlx.DB) {
	sats, _ := QuerySats(db)

	home := pages.HomePage()
	login := pages.LoginPage()
	dashboard := pages.DashboardPage(sats)
	plan := pages.PlanPage(sats)
	schedule := pages.SchedulePage()

	filesrv := http.FileServer(http.Dir("assets"))

	router.Handle("/images/*", filesrv)
	router.Handle("/css/output.css", filesrv)
	router.Handle("/assets/*", http.StripPrefix("/assets/", filesrv))

	router.Handle("/", templ.Handler(home))
	router.Handle("/auth/login", templ.Handler(login))
	router.Handle("/dashboard", templ.Handler(dashboard))
	router.Handle("/plan", templ.Handler(plan))
	router.Handle("/schedule", templ.Handler(schedule))

	router.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)
	})
}
