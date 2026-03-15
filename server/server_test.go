package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func executeRequest(req *http.Request, s *Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected int, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestHomeUnauthenticated(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		t.Errorf("Failed to load env file\n")
	}

	key := os.Getenv("SERVER_SECRET_KEY")
	if key == "" {
		t.Errorf("Failed to load env var\n")
	}

	db, err := sqlx.Connect("sqlite3", "./vertex.db")
	if err != nil {
		t.Errorf("Failed to create db connection\n")
	}

	defer db.Close()
	srv := NewServer(key, db)
	srv.ServeAssets()
	srv.MountHandlers()
	req, _ := http.NewRequest("GET", "/", nil)
	response := executeRequest(req, &srv)
	checkResponseCode(t, http.StatusSeeOther, response.Code)
	require.Equal(t, "<a href=\"/auth/login\">See Other</a>.\n\n", response.Body.String())
}

func TestHomeAuthenticated(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		t.Errorf("Failed to load env file\n")
	}

	key := os.Getenv("SERVER_SECRET_KEY")
	if key == "" {
		t.Errorf("Failed to load env var\n")
	}

	db, err := sqlx.Connect("sqlite3", "./vertex.db")
	if err != nil {
		t.Errorf("Failed to create db connection\n")
	}

	defer db.Close()
	srv := NewServer(key, db)
	srv.ServeAssets()
	srv.MountHandlers()
	req, _ := http.NewRequest("GET", "/", nil)
	response := executeRequest(req, &srv)
	checkResponseCode(t, http.StatusOK, response.Code)
	require.Equal(t, "", response.Body.String())
}
