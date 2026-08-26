package auth

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
			name:    "email already exists",
			err:     fmt.Errorf("register user: %w", ErrEmailAlreadyExists),
			status:  http.StatusConflict,
			code:    "EMAIL_ALREADY_EXISTS",
			message: "email already exists",
		},
		{
			name:    "phone number already exists",
			err:     ErrPhoneNumberAlreadyExists,
			status:  http.StatusConflict,
			code:    "PHONE_NUMBER_ALREADY_EXISTS",
			message: "phone number already exists",
		},
		{
			name:    "invalid credentials",
			err:     ErrInvalidCredentials,
			status:  http.StatusUnauthorized,
			code:    "INVALID_CREDENTIALS",
			message: "invalid credentials",
		},
		{
			name:    "invalid email",
			err:     ErrInvalidEmail,
			status:  http.StatusBadRequest,
			code:    "INVALID_EMAIL",
			message: "invalid email",
		},
		{
			name:    "invalid password",
			err:     ErrPasswordMissingSpecial,
			status:  http.StatusBadRequest,
			code:    "INVALID_PASSWORD",
			message: "password must be 8 to 40 characters and contain a number and special character",
		},
		{
			name:    "name too long",
			err:     ErrNameTooLong,
			status:  http.StatusBadRequest,
			code:    "INVALID_NAME",
			message: "name must be at most 25 characters",
		},
		{
			name:    "phone number too long",
			err:     ErrPhoneNumberTooLong,
			status:  http.StatusBadRequest,
			code:    "INVALID_PHONE_NUMBER",
			message: "phone number must be at most 15 characters",
		},
		{
			name:    "invalid request",
			err:     ErrInvalidRequest,
			status:  http.StatusBadRequest,
			code:    "INVALID_REQUEST",
			message: "invalid request",
		},
		{
			name:    "unauthorized",
			err:     ErrUnauthorized,
			status:  http.StatusUnauthorized,
			code:    "UNAUTHORIZED",
			message: "unauthorized",
		},
		{
			name:    "invalid token",
			err:     ErrInvalidToken,
			status:  http.StatusUnauthorized,
			code:    "INVALID_TOKEN",
			message: "invalid token",
		},
		{
			name:    "expired token",
			err:     ErrExpiredToken,
			status:  http.StatusUnauthorized,
			code:    "EXPIRED_TOKEN",
			message: "token expired",
		},
		{
			name:    "unexpected error",
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
				t.Fatalf(
					"mapError() = %#v, want status=%d code=%q message=%q",
					got,
					tt.status,
					tt.code,
					tt.message,
				)
			}
		})
	}
}
