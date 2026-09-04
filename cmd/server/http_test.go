package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Takamasa045/Yokai-Finder-MCP/internal/handler"
)

func TestCheckListenAddr(t *testing.T) {
	if err := checkListenAddr("127.0.0.1:8080", ""); err != nil {
		t.Fatalf("loopback should be allowed: %v", err)
	}
	if err := checkListenAddr("0.0.0.0:8080", ""); err == nil {
		t.Fatal("public bind without token should fail")
	}
	if err := checkListenAddr("0.0.0.0:8080", "short"); err == nil {
		t.Fatal("short token should fail on public bind")
	}
	if err := checkListenAddr("0.0.0.0:8080", "sixteen-chars-ok"); err != nil {
		t.Fatalf("public bind with long token should pass: %v", err)
	}
}

func TestHTTPGuardsAuth(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	h := withHTTPGuards(inner, "sixteen-chars-ok", false)

	unauth := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, unauth)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing bearer: got %d", rec.Code)
	}

	wrong := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	wrong.Header.Set("Authorization", "Bearer nope-nope-nope-nope")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, wrong)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong bearer: got %d", rec.Code)
	}

	okReq := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	okReq.Header.Set("Authorization", "Bearer sixteen-chars-ok")
	okReq.RemoteAddr = "10.0.0.1:1234"
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, okReq)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("valid bearer: got %d", rec.Code)
	}
}

func TestIPLimiter(t *testing.T) {
	l := newIPLimiter(2, time.Minute, 16)
	if !l.allow("1.1.1.1") || !l.allow("1.1.1.1") {
		t.Fatal("first two hits should pass")
	}
	if l.allow("1.1.1.1") {
		t.Fatal("third hit should be limited")
	}
	if !l.allow("2.2.2.2") {
		t.Fatal("other IP should pass")
	}
}

func TestHealthzAndLoopbackGuard(t *testing.T) {
	h := newHTTPHandler(newServer(handler.New(nil, nil)), "", true)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if strings.TrimSpace(string(body)) != "ok" {
		t.Fatalf("healthz body %q", body)
	}

	blocked := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	blocked.Host = "example.com"
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, blocked)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("non-loopback host should be forbidden, got %d", rec.Code)
	}
}
