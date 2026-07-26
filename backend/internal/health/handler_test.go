package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLive(t *testing.T) {
	handler := NewHandler(func(context.Context) error { return nil })
	recorder := httptest.NewRecorder()
	handler.Live(recorder, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	assertResponse(t, recorder, http.StatusOK, "{\"status\":\"ok\"}\n")
}

func TestReady(t *testing.T) {
	tests := []struct {
		name   string
		ping   PingFunc
		status int
		body   string
	}{
		{name: "database available", ping: func(context.Context) error { return nil }, status: http.StatusOK, body: "{\"status\":\"ok\"}\n"},
		{name: "database unavailable", ping: func(context.Context) error { return errors.New("connection refused") }, status: http.StatusServiceUnavailable, body: "{\"status\":\"unavailable\"}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.ping)
			recorder := httptest.NewRecorder()
			handler.Ready(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
			assertResponse(t, recorder, tt.status, tt.body)
		})
	}
}

func assertResponse(t *testing.T, recorder *httptest.ResponseRecorder, status int, body string) {
	t.Helper()
	if recorder.Code != status {
		t.Fatalf("status = %d, want %d", recorder.Code, status)
	}
	if recorder.Body.String() != body {
		t.Fatalf("body = %q, want %q", recorder.Body.String(), body)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
}
