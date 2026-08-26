package auth

import (
	"errors"
	"net/http"

	"immera/internal/platform/httpx"
)

func mapError(err error) httpx.MappedError {
	switch {
	case errors.Is(err, ErrEmailAlreadyExists):
		return httpx.MappedError{
			Status:  http.StatusConflict,
			Code:    "EMAIL_ALREADY_EXISTS",
			Message: "email already exists",
		}
	case errors.Is(err, ErrPhoneNumberAlreadyExists):
		return httpx.MappedError{
			Status:  http.StatusConflict,
			Code:    "PHONE_NUMBER_ALREADY_EXISTS",
			Message: "phone number already exists",
		}
	case errors.Is(err, ErrInvalidCredentials):
		return httpx.MappedError{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_CREDENTIALS",
			Message: "invalid credentials",
		}
	case errors.Is(err, ErrInvalidEmail):
		return httpx.MappedError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_EMAIL",
			Message: "invalid email",
		}
	case errors.Is(err, ErrInvalidRequest):
		return httpx.MappedError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_REQUEST",
			Message: "invalid request",
		}
	case errors.Is(err, ErrUnauthorized):
		return httpx.MappedError{
			Status:  http.StatusUnauthorized,
			Code:    "UNAUTHORIZED",
			Message: "unauthorized",
		}
	case errors.Is(err, ErrInvalidToken):
		return httpx.MappedError{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_TOKEN",
			Message: "invalid token",
		}
	case errors.Is(err, ErrExpiredToken):
		return httpx.MappedError{
			Status:  http.StatusUnauthorized,
			Code:    "EXPIRED_TOKEN",
			Message: "token expired",
		}
	default:
		return httpx.MappedError{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		}
	}
}

func (h *Handler) writeMappedError(
	w http.ResponseWriter,
	err error,
	logMessage string,
	logAttrs ...any,
) {
	mapped := mapError(err)

	if mapped.Status == http.StatusInternalServerError {
		attrs := httpx.AppendLogError(logAttrs, err)
		h.log.Error(logMessage, attrs...)
	}

	if writeErr := httpx.WriteError(w, mapped.Status, mapped.Code, mapped.Message); writeErr != nil {
		attrs := httpx.AppendLogError(logAttrs, writeErr)
		h.log.Error("failed to write error response", attrs...)
	}
}
