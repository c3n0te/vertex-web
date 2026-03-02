package main

import (
	"net/http"

	"vertex/ui/pages"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
)

func MountRoutes(router chi.Router) {
	home := pages.HomePage()
	login := pages.LoginPage()

	filesrv := http.FileServer(http.Dir("assets"))

	router.Handle("/images/*", filesrv)
	router.Handle("/css/output.css", filesrv)
	router.Handle("/", templ.Handler(home))
	router.Handle("/auth/login", templ.Handler(login))
}
