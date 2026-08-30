package user

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	httpserver "immera/internal/platform/http"
	"immera/internal/platform/httpx"
)

const userHandlerTestJWTSecret = "test-secret-that-is-at-least-32-bytes"

type stubUserService struct {
	user                 User
	getByIDUserID        uuid.UUID
	getByIDCalled        bool
	getByIDErr           error
	settings             UserSettings
	settingsUserID       uuid.UUID
	settingsCalled       bool
	settingsErr          error
	updateUserID         uuid.UUID
	updateParams         UpdateUserParams
	updateCalled         bool
	updateErr            error
	updateSettingsUserID uuid.UUID
	updateSettingsParams UpdateSettingsParams
	updateSettingsCalled bool
	updateSettingsErr    error
}

func (s *stubUserService) GetByID(_ context.Context, userID uuid.UUID) (User, error) {
	s.getByIDCalled = true
	s.getByIDUserID = userID
	return s.user, s.getByIDErr
}

func (s *stubUserService) GetUserSettings(
	_ context.Context,
	userID uuid.UUID,
) (UserSettings, error) {
	s.settingsCalled = true
	s.settingsUserID = userID
	return s.settings, s.settingsErr
}

func (s *stubUserService) UpdateUser(
	_ context.Context,
	userID uuid.UUID,
	params UpdateUserParams,
) error {
	s.updateCalled = true
	s.updateUserID = userID
	s.updateParams = params
	return s.updateErr
}

func (s *stubUserService) UpdateSettings(
	_ context.Context,
	userID uuid.UUID,
	params UpdateSettingsParams,
) error {
	s.updateSettingsCalled = true
	s.updateSettingsUserID = userID
	s.updateSettingsParams = params
	return s.updateSettingsErr
}

func TestGetMeUsesAuthenticatedUserID(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	name := "Jane Doe"
	phoneNumber := "+48123456789"
	service := &stubUserService{
		user: User{
			ID:          userID,
			Name:        &name,
			Email:       "jane@example.com",
			PhoneNumber: &phoneNumber,
		},
	}
	router := newProtectedUserTestRouter(service)
	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		newAuthenticatedUserRequest(t, http.MethodGet, "/api/v1/users/me", nil, userID),
	)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !service.getByIDCalled || service.getByIDUserID != userID {
		t.Fatalf("GetByID called = %t with %s, want %s", service.getByIDCalled, service.getByIDUserID, userID)
	}

	var body userByIDResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ID != userID || body.Email != "jane@example.com" || body.Name == nil || *body.Name != name ||
		body.PhoneNumber == nil || *body.PhoneNumber != phoneNumber {
		t.Fatalf("response = %#v", body)
	}
}

func TestGetMySettingsUsesAuthenticatedUserID(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	settingsID := uuid.New()
	service := &stubUserService{
		settings: UserSettings{
			ID:              settingsID,
			DefaultLanguage: "en",
			Theme:           "dark",
		},
	}
	router := newProtectedUserTestRouter(service)
	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		newAuthenticatedUserRequest(t, http.MethodGet, "/api/v1/users/me/settings", nil, userID),
	)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !service.settingsCalled || service.settingsUserID != userID {
		t.Fatalf(
			"GetUserSettings called = %t with %s, want %s",
			service.settingsCalled,
			service.settingsUserID,
			userID,
		)
	}

	var body userSettingsResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ID != settingsID || body.DefaultLanguage != "en" || body.Theme != "dark" {
		t.Fatalf("response = %#v", body)
	}
}

func TestUpdateMeUsesAuthenticatedUserID(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	service := &stubUserService{}
	router := newProtectedUserTestRouter(service)
	response := httptest.NewRecorder()
	body := strings.NewReader(`{"name":"Jane Doe","email":"jane@example.com","phone_number":"+48123456789"}`)

	router.ServeHTTP(
		response,
		newAuthenticatedUserRequest(t, http.MethodPut, "/api/v1/users/me", body, userID),
	)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if !service.updateCalled || service.updateUserID != userID {
		t.Fatalf("UpdateUser called = %t with %s, want %s", service.updateCalled, service.updateUserID, userID)
	}
	want := UpdateUserParams{
		Name:        "Jane Doe",
		Email:       "jane@example.com",
		PhoneNumber: "+48123456789",
	}
	if service.updateParams != want {
		t.Fatalf("UpdateUser params = %#v, want %#v", service.updateParams, want)
	}
}

func TestUpdateMeRejectsInvalidRequestBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: `{"name":`},
		{
			name: "client-supplied user ID",
			body: `{"name":"Jane","email":"jane@example.com","phone_number":"+48123456789",` +
				`"user_id":"00000000-0000-0000-0000-000000000000"}`,
		},
		{
			name: "multiple JSON values",
			body: `{"name":"Jane","email":"jane@example.com","phone_number":"+48123456789"} {}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userID := uuid.New()
			service := &stubUserService{}
			router := newProtectedUserTestRouter(service)
			request := newAuthenticatedUserRequest(
				t,
				http.MethodPut,
				"/api/v1/users/me",
				strings.NewReader(tt.body),
				userID,
			)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
			assertUserErrorResponse(t, response, "INVALID_REQUEST", "invalid request")
			if service.updateCalled {
				t.Fatal("service was called for an invalid update request")
			}
		})
	}
}

func TestUpdateMySettingsUsesAuthenticatedUserID(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	service := &stubUserService{}
	router := newProtectedUserTestRouter(service)
	response := httptest.NewRecorder()
	body := strings.NewReader(`{"default_language":"en","theme":"dark"}`)

	router.ServeHTTP(
		response,
		newAuthenticatedUserRequest(t, http.MethodPut, "/api/v1/users/me/settings", body, userID),
	)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if !service.updateSettingsCalled || service.updateSettingsUserID != userID {
		t.Fatalf(
			"UpdateSettings called = %t with %s, want %s",
			service.updateSettingsCalled,
			service.updateSettingsUserID,
			userID,
		)
	}
	want := UpdateSettingsParams{DefaultLanguage: "en", Theme: "dark"}
	if service.updateSettingsParams != want {
		t.Fatalf("UpdateSettings params = %#v, want %#v", service.updateSettingsParams, want)
	}
}

func TestUpdateMySettingsRejectsInvalidRequestBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: `{"default_language":`},
		{
			name: "unknown field",
			body: `{"default_language":"en","theme":"dark","user_id":` +
				`"00000000-0000-0000-0000-000000000000"}`,
		},
		{
			name: "multiple JSON values",
			body: `{"default_language":"en","theme":"dark"} {}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userID := uuid.New()
			service := &stubUserService{}
			router := newProtectedUserTestRouter(service)
			request := newAuthenticatedUserRequest(
				t,
				http.MethodPut,
				"/api/v1/users/me/settings",
				strings.NewReader(tt.body),
				userID,
			)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
			assertUserErrorResponse(t, response, "INVALID_REQUEST", "invalid request")
			if service.updateSettingsCalled {
				t.Fatal("service was called for an invalid settings request")
			}
		})
	}
}

func TestCurrentUserHandlersTreatMissingContextAsInternalError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
		body   io.Reader
		serve  func(*Handler, http.ResponseWriter, *http.Request)
	}{
		{
			name:   "get current user",
			method: http.MethodGet,
			path:   "/users/me",
			serve:  (*Handler).GetMe,
		},
		{
			name:   "get current user settings",
			method: http.MethodGet,
			path:   "/users/me/settings",
			serve:  (*Handler).GetUserSettings,
		},
		{
			name:   "update current user",
			method: http.MethodPut,
			path:   "/users/me",
			body:   strings.NewReader(`{"name":"Jane","email":"jane@example.com","phone_number":"+48123456789"}`),
			serve:  (*Handler).UpdateUser,
		},
		{
			name:   "update current user settings",
			method: http.MethodPut,
			path:   "/users/me/settings",
			body:   strings.NewReader(`{"default_language":"en","theme":"dark"}`),
			serve:  (*Handler).UpdateUserSettings,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := &stubUserService{}
			handler := NewHandler(service, discardUserTestLogger())
			request := httptest.NewRequest(tt.method, tt.path, tt.body)
			response := httptest.NewRecorder()

			tt.serve(handler, response, request)

			if response.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
			}
			assertUserErrorResponse(t, response, "INTERNAL_ERROR", "internal server error")
			if service.getByIDCalled || service.settingsCalled || service.updateCalled || service.updateSettingsCalled {
				t.Fatal("service was called without an authenticated user ID")
			}
		})
	}
}

func TestProtectedUserRoutesRequireAuthentication(t *testing.T) {
	t.Parallel()

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/users/me"},
		{method: http.MethodGet, path: "/api/v1/users/me/settings"},
		{method: http.MethodPut, path: "/api/v1/users/me"},
		{method: http.MethodPut, path: "/api/v1/users/me/settings"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			t.Parallel()

			service := &stubUserService{}
			router := newProtectedUserTestRouter(service)
			response := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, nil)

			router.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
			assertUserErrorResponse(t, response, "UNAUTHORIZED", "unauthorized")
			if service.getByIDCalled || service.settingsCalled || service.updateCalled || service.updateSettingsCalled {
				t.Fatal("service was called for an unauthenticated request")
			}
		})
	}
}

func TestLegacyExplicitIDRoutesAreNotRegistered(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/users/" + userID.String()},
		{method: http.MethodGet, path: "/api/v1/users/" + userID.String() + "/settings"},
		{method: http.MethodPut, path: "/api/v1/users/" + userID.String()},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			t.Parallel()

			service := &stubUserService{}
			router := newProtectedUserTestRouter(service)
			request := newAuthenticatedUserRequest(t, tt.method, tt.path, nil, userID)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
			}
			if service.getByIDCalled || service.settingsCalled || service.updateCalled || service.updateSettingsCalled {
				t.Fatal("service was called through a legacy explicit-ID route")
			}
		})
	}
}

func newProtectedUserTestRouter(service userService) http.Handler {
	handler := NewHandler(service, discardUserTestLogger())
	return httpserver.NewRouter(
		discardUserTestLogger(),
		nil,
		[]byte(userHandlerTestJWTSecret),
		nil,
		nil,
		[]httpserver.RouteRegistrar{handler.ProtectedRoutes},
	)
}

func newAuthenticatedUserRequest(
	t *testing.T,
	method string,
	path string,
	body io.Reader,
	userID uuid.UUID,
) *http.Request {
	t.Helper()

	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(userHandlerTestJWTSecret))
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}

	request := httptest.NewRequest(method, path, body)
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func assertUserErrorResponse(
	t *testing.T,
	response *httptest.ResponseRecorder,
	code string,
	message string,
) {
	t.Helper()

	var body httpx.ErrorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error.Code != code || body.Error.Message != message {
		t.Fatalf("error response = %#v, want code=%q message=%q", body, code, message)
	}
}

func discardUserTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
}
