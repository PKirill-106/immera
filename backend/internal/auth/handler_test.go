package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"immera/internal/platform/httpx"
)

type stubAuthService struct {
	registerErr error
	registered  RegisterDTO
	called      bool
}

func (s *stubAuthService) Register(_ context.Context, registration RegisterDTO) error {
	s.called = true
	s.registered = registration
	return s.registerErr
}

func TestRegisterRejectsInvalidRequestBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: `{"email":`},
		{name: "unknown field", body: `{"email":"jane@example.com","password":"password","unknown":true}`},
		{name: "multiple JSON values", body: `{"email":"jane@example.com","password":"password"} {}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := &stubAuthService{}
			handler := newTestHandler(service)
			request := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			handler.Register(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
			if service.called {
				t.Fatal("service was called for an invalid request body")
			}

			assertErrorResponse(t, response, "INVALID_REQUEST", "invalid request")
		})
	}
}

func TestRegisterMapsServiceError(t *testing.T) {
	t.Parallel()

	service := &stubAuthService{
		registerErr: fmt.Errorf("register user: %w", ErrEmailAlreadyExists),
	}
	handler := newTestHandler(service)
	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(`{"email":"jane@example.com","password":"password"}`),
	)
	response := httptest.NewRecorder()

	handler.Register(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
	assertErrorResponse(t, response, "EMAIL_ALREADY_EXISTS", "email already exists")
}

func TestRegisterReturnsCreatedWithoutBody(t *testing.T) {
	t.Parallel()

	service := &stubAuthService{}
	handler := newTestHandler(service)
	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/register",
		strings.NewReader(`{"name":"Jane","email":"jane@example.com","phone_number":"+48123456789","password":"password"}`),
	)
	response := httptest.NewRecorder()

	handler.Register(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if response.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty body", response.Body.String())
	}
	if !service.called {
		t.Fatal("service was not called")
	}
	if service.registered.Name == nil || *service.registered.Name != "Jane" {
		t.Fatalf("registered name = %v, want Jane", service.registered.Name)
	}
	if service.registered.PhoneNumber == nil || *service.registered.PhoneNumber != "+48123456789" {
		t.Fatalf("registered phone number = %v, want +48123456789", service.registered.PhoneNumber)
	}
}

func newTestHandler(service authService) *Handler {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewHandler(service, log)
}

func assertErrorResponse(t *testing.T, response *httptest.ResponseRecorder, code, message string) {
	t.Helper()

	var body httpx.ErrorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != code || body.Error.Message != message {
		t.Fatalf("error response = %#v, want code=%q message=%q", body, code, message)
	}
}
