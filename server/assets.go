package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func ServeAssets(router chi.Router) {
	filesrv := http.FileServer(http.Dir("assets"))
	router.Handle("/images/*", filesrv)
	router.Handle("/css/output.css", filesrv)
	router.Handle("/assets/*", http.StripPrefix("/assets/", filesrv))
}
