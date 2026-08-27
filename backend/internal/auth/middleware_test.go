package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"immera/internal/platform/httpx"
)

func TestMiddlewareRejectsUnauthorizedRequests(t *testing.T) {
	t.Parallel()

	secret := []byte("test-secret-that-is-at-least-32-bytes")
	userID := uuid.New()
	now := time.Now()

	validClaims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
	}

	tests := []struct {
		name          string
		authorization string
	}{
		{name: "missing Authorization header"},
		{name: "empty Authorization header", authorization: ""},
		{name: "wrong authentication scheme", authorization: "Basic credentials"},
		{name: "Bearer without token", authorization: "Bearer "},
		{name: "malformed JWT", authorization: "Bearer not-a-jwt"},
		{
			name: "invalid signature",
			authorization: "Bearer " + signMiddlewareTestToken(
				t,
				jwt.SigningMethodHS256,
				[]byte("different-secret-that-is-at-least-32-bytes"),
				validClaims,
			),
		},
		{
			name: "expired JWT",
			authorization: "Bearer " + signMiddlewareTestToken(
				t,
				jwt.SigningMethodHS256,
				secret,
				jwt.RegisteredClaims{
					Subject:   userID.String(),
					IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
					ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
				},
			),
		},
		{
			name: "non-HS256 signing method",
			authorization: "Bearer " + signMiddlewareTestToken(
				t,
				jwt.SigningMethodHS384,
				secret,
				validClaims,
			),
		},
		{
			name: "missing subject",
			authorization: "Bearer " + signMiddlewareTestToken(
				t,
				jwt.SigningMethodHS256,
				secret,
				jwt.RegisteredClaims{
					IssuedAt:  jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				},
			),
		},
		{
			name: "empty subject",
			authorization: "Bearer " + signMiddlewareTestToken(
				t,
				jwt.SigningMethodHS256,
				secret,
				jwt.MapClaims{
					"sub": "",
					"iat": now.Unix(),
					"exp": now.Add(time.Hour).Unix(),
				},
			),
		},
		{
			name: "invalid UUID subject",
			authorization: "Bearer " + signMiddlewareTestToken(
				t,
				jwt.SigningMethodHS256,
				secret,
				jwt.RegisteredClaims{
					Subject:   "not-a-uuid",
					IssuedAt:  jwt.NewNumericDate(now),
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				},
			),
		},
		{
			name: "missing expiration",
			authorization: "Bearer " + signMiddlewareTestToken(
				t,
				jwt.SigningMethodHS256,
				secret,
				jwt.RegisteredClaims{
					Subject:  userID.String(),
					IssuedAt: jwt.NewNumericDate(now),
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var logs bytes.Buffer
			log := slog.New(slog.NewJSONHandler(&logs, nil))
			nextCalled := false
			next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				nextCalled = true
			})
			handler := Middleware(secret, log)(next)
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.name != "missing Authorization header" {
				request.Header.Set("Authorization", tt.authorization)
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
			assertMiddlewareUnauthorizedResponse(t, response)
			if nextCalled {
				t.Fatal("next handler was called for an unauthorized request")
			}
			if strings.Contains(logs.String(), `"level":"ERROR"`) {
				t.Fatalf("authentication failure was logged as ERROR: %s", logs.String())
			}
		})
	}
}

func TestMiddlewareStoresAuthenticatedUserIDAndCallsNext(t *testing.T) {
	t.Parallel()

	secret := []byte("test-secret-that-is-at-least-32-bytes")
	wantUserID := uuid.New()
	token, err := generateAccessToken(wantUserID, secret, time.Hour)
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			t.Fatal("authenticated user ID was not stored in request context")
		}
		if userID != wantUserID {
			t.Fatalf("context user ID = %s, want %s", userID, wantUserID)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := Middleware(secret, discardMiddlewareTestLogger())(next)
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if !nextCalled {
		t.Fatal("next handler was not called for a valid access token")
	}
}

func TestUserIDFromContextWithoutAuthenticatedUser(t *testing.T) {
	t.Parallel()

	userID, ok := UserIDFromContext(context.Background())
	if ok {
		t.Fatalf("UserIDFromContext() = (%s, true), want (_, false)", userID)
	}
}

func TestMiddlewareLogsResponseWriteFailure(t *testing.T) {
	t.Parallel()

	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := Middleware(
		[]byte("test-secret-that-is-at-least-32-bytes"),
		log,
	)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler was called for an unauthorized request")
	}))

	handler.ServeHTTP(
		&failingMiddlewareResponseWriter{header: make(http.Header)},
		httptest.NewRequest(http.MethodGet, "/protected", nil),
	)

	if !strings.Contains(logs.String(), `"level":"ERROR"`) ||
		!strings.Contains(logs.String(), "failed to write unauthorized response") {
		t.Fatalf("response write failure was not logged as ERROR: %s", logs.String())
	}
}

func signMiddlewareTestToken(
	t *testing.T,
	method jwt.SigningMethod,
	secret []byte,
	claims jwt.Claims,
) string {
	t.Helper()

	token, err := jwt.NewWithClaims(method, claims).SignedString(secret)
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}

	return token
}

func assertMiddlewareUnauthorizedResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	const wantJSON = `{"error":{"code":"UNAUTHORIZED","message":"unauthorized"}}` + "\n"
	if response.Body.String() != wantJSON {
		t.Fatalf("unauthorized response body = %q, want %q", response.Body.String(), wantJSON)
	}

	var body httpx.ErrorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode unauthorized response: %v", err)
	}
	if body.Error.Code != "UNAUTHORIZED" || body.Error.Message != "unauthorized" {
		t.Fatalf("unauthorized response = %#v", body)
	}
}

func discardMiddlewareTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
}

type failingMiddlewareResponseWriter struct {
	header http.Header
}

func (w *failingMiddlewareResponseWriter) Header() http.Header {
	return w.header
}

func (*failingMiddlewareResponseWriter) WriteHeader(int) {}

func (*failingMiddlewareResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
