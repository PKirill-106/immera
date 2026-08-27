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
	registerErr    error
	registered     RegisterDTO
	registerCalled bool
	loginErr       error
	loginParams    LoginParams
	loginTokens    TokenPair
	loginCalled    bool
	refreshErr     error
	refreshParams  RefreshParams
	refreshTokens  TokenPair
	refreshCalled  bool
	logoutErr      error
	logoutParams   LogoutParams
	logoutCalled   bool
}

func (s *stubAuthService) Register(_ context.Context, registration RegisterDTO) error {
	s.registerCalled = true
	s.registered = registration
	return s.registerErr
}

func (s *stubAuthService) Login(_ context.Context, params LoginParams) (TokenPair, error) {
	s.loginCalled = true
	s.loginParams = params
	return s.loginTokens, s.loginErr
}

func (s *stubAuthService) Refresh(_ context.Context, params RefreshParams) (TokenPair, error) {
	s.refreshCalled = true
	s.refreshParams = params
	return s.refreshTokens, s.refreshErr
}

func (s *stubAuthService) Logout(_ context.Context, params LogoutParams) error {
	s.logoutCalled = true
	s.logoutParams = params
	return s.logoutErr
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
			if service.registerCalled {
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
	if !service.registerCalled {
		t.Fatal("service was not called")
	}
	if service.registered.Name == nil || *service.registered.Name != "Jane" {
		t.Fatalf("registered name = %v, want Jane", service.registered.Name)
	}
	if service.registered.PhoneNumber == nil || *service.registered.PhoneNumber != "+48123456789" {
		t.Fatalf("registered phone number = %v, want +48123456789", service.registered.PhoneNumber)
	}
}

func TestLoginRejectsInvalidRequestBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: `{"email":`},
		{name: "unknown field", body: `{"email":"jane@example.com","password":"password1!","unknown":true}`},
		{name: "multiple JSON values", body: `{"email":"jane@example.com","password":"password1!"} {}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := &stubAuthService{}
			handler := newTestHandler(service)
			request := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.body))
			response := httptest.NewRecorder()

			handler.Login(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
			if service.loginCalled {
				t.Fatal("service was called for an invalid request body")
			}

			assertErrorResponse(t, response, "INVALID_REQUEST", "invalid request")
		})
	}
}

func TestLoginMapsInvalidCredentials(t *testing.T) {
	t.Parallel()

	service := &stubAuthService{loginErr: ErrInvalidCredentials}
	handler := newTestHandler(service)
	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(`{"email":"jane@example.com","password":"wrong-password"}`),
	)
	response := httptest.NewRecorder()

	handler.Login(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	assertErrorResponse(t, response, "INVALID_CREDENTIALS", "invalid credentials")
}

func TestLoginReturnsTokenPair(t *testing.T) {
	t.Parallel()

	service := &stubAuthService{
		loginTokens: TokenPair{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
		},
	}
	handler := newTestHandler(service)
	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		strings.NewReader(`{"email":"jane@example.com","password":"password1!"}`),
	)
	response := httptest.NewRecorder()

	handler.Login(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
	if !service.loginCalled {
		t.Fatal("service was not called")
	}
	if service.loginParams.Email != "jane@example.com" || service.loginParams.Password != "password1!" {
		t.Fatalf("login params = %#v, want submitted credentials", service.loginParams)
	}

	var body tokenPairResponseDTO
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.AccessToken != "access-token" || body.RefreshToken != "refresh-token" {
		t.Fatalf("response = %#v, want returned token pair", body)
	}
}

func TestRefreshRejectsEmptyToken(t *testing.T) {
	t.Parallel()

	service := &stubAuthService{}
	handler := newTestHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(`{"refresh_token":""}`))
	response := httptest.NewRecorder()

	handler.Refresh(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if service.refreshCalled {
		t.Fatal("service was called for an empty refresh token")
	}
	assertErrorResponse(t, response, "INVALID_REQUEST", "invalid request")
}

func TestRefreshReturnsRotatedTokenPair(t *testing.T) {
	t.Parallel()

	service := &stubAuthService{
		refreshTokens: TokenPair{AccessToken: "new-access", RefreshToken: "new-refresh"},
	}
	handler := newTestHandler(service)
	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/refresh",
		strings.NewReader(`{"refresh_token":"old-refresh"}`),
	)
	response := httptest.NewRecorder()

	handler.Refresh(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if service.refreshParams.RefreshToken != "old-refresh" {
		t.Fatalf("refresh token = %q, want old-refresh", service.refreshParams.RefreshToken)
	}

	var body tokenPairResponseDTO
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.AccessToken != "new-access" || body.RefreshToken != "new-refresh" {
		t.Fatalf("response = %#v, want rotated token pair", body)
	}
}

func TestLogoutReturnsNoContent(t *testing.T) {
	t.Parallel()

	service := &stubAuthService{}
	handler := newTestHandler(service)
	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/logout",
		strings.NewReader(`{"refresh_token":"refresh-token"}`),
	)
	response := httptest.NewRecorder()

	handler.Logout(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if service.logoutParams.RefreshToken != "refresh-token" {
		t.Fatalf("logout token = %q, want refresh-token", service.logoutParams.RefreshToken)
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
