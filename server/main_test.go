package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Index(w http.ResponseWriter, r *http.Request) {

}

func TestIndex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(Index))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
}
