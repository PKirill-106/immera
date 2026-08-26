package user

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestMapError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		status  int
		code    string
		message string
	}{
		{
			name:    "invalid user ID",
			err:     fmt.Errorf("wrapped: %w", ErrInvalidUserID),
			status:  http.StatusBadRequest,
			code:    "INVALID_USER_ID",
			message: "invalid user id",
		},
		{
			name:    "invalid request",
			err:     ErrInvalidRequest,
			status:  http.StatusBadRequest,
			code:    "INVALID_REQUEST",
			message: "invalid request",
		},
		{
			name:    "invalid user data",
			err:     ErrInvalidUserData,
			status:  http.StatusBadRequest,
			code:    "INVALID_USER_DATA",
			message: "invalid user data",
		},
		{
			name:    "user not found",
			err:     ErrUserNotFound,
			status:  http.StatusNotFound,
			code:    "USER_NOT_FOUND",
			message: "user not found",
		},
		{
			name:    "email conflict",
			err:     ErrEmailAlreadyExists,
			status:  http.StatusConflict,
			code:    "EMAIL_ALREADY_EXISTS",
			message: "email already exists",
		},
		{
			name:    "phone conflict",
			err:     ErrPhoneNumberAlreadyExists,
			status:  http.StatusConflict,
			code:    "PHONE_NUMBER_ALREADY_EXISTS",
			message: "phone number already exists",
		},
		{
			name:    "generic user conflict",
			err:     ErrUserConflict,
			status:  http.StatusConflict,
			code:    "USER_CONFLICT",
			message: "user conflicts with existing data",
		},
		{
			name:    "unknown error",
			err:     errors.New("database unavailable"),
			status:  http.StatusInternalServerError,
			code:    "INTERNAL_ERROR",
			message: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := mapError(tt.err)
			if got.Status != tt.status || got.Code != tt.code || got.Message != tt.message {
				t.Fatalf("mapError() = %#v, want status=%d code=%q message=%q", got, tt.status, tt.code, tt.message)
			}
		})
	}
}
