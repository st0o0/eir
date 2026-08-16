package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthAddr_PortOnly(t *testing.T) {
	if got := healthAddr(":9550"); got != "localhost:9550" {
		t.Errorf("healthAddr(':9550') = %q, want 'localhost:9550'", got)
	}
}

func TestHealthAddr_Wildcard(t *testing.T) {
	if got := healthAddr("0.0.0.0:9550"); got != "localhost:9550" {
		t.Errorf("healthAddr('0.0.0.0:9550') = %q, want 'localhost:9550'", got)
	}
}

func TestHealthAddr_IPv6Wildcard(t *testing.T) {
	if got := healthAddr("[::]:9550"); got != "localhost:9550" {
		t.Errorf("healthAddr('[::]:9550') = %q, want 'localhost:9550'", got)
	}
}

func TestHealthAddr_Qualified(t *testing.T) {
	if got := healthAddr("10.0.0.1:9550"); got != "10.0.0.1:9550" {
		t.Errorf("healthAddr('10.0.0.1:9550') = %q, want '10.0.0.1:9550'", got)
	}
}

func TestDoHealthcheck_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"ok"}`)
	}))
	defer srv.Close()

	addr := srv.Listener.Addr().String()
	if code := doHealthcheck(addr); code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

func TestDoHealthcheck_ServiceUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"status":"error"}`)
	}))
	defer srv.Close()

	addr := srv.Listener.Addr().String()
	if code := doHealthcheck(addr); code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
}

func TestDoHealthcheck_ConnectionRefused(t *testing.T) {
	if code := doHealthcheck("localhost:1"); code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
}
