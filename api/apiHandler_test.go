package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiHandlerNoRepo(t *testing.T) {
	conf := Conf{}
	server := httptest.NewServer(http.HandlerFunc(conf.ApiHandler))
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("An error occured while making the request: %v\n", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Unexpected status %d, expected %d\n", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestApiHandlerNoToken(t *testing.T) {
	conf := Conf{}
	server := httptest.NewServer(http.HandlerFunc(conf.ApiHandler))
	resp, err := http.Get(server.URL + "?repo=evermax/stargraph")
	if err != nil {
		t.Fatalf("An error occured while making the request: %v\n", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Unexpected status %d, expected %d\n", resp.StatusCode, http.StatusBadRequest)
	}
}
