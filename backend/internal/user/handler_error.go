package user

import (
	"errors"
	"net/http"

	"immera/internal/platform/httpx"
)

func mapError(err error) httpx.MappedError {
	switch {
	case errors.Is(err, ErrInvalidUserID):
		return httpx.MappedError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_USER_ID",
			Message: "invalid user id",
		}
	case errors.Is(err, ErrInvalidRequest):
		return httpx.MappedError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_REQUEST",
			Message: "invalid request",
		}
	case errors.Is(err, ErrInvalidUserData):
		return httpx.MappedError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_USER_DATA",
			Message: "invalid user data",
		}
	case errors.Is(err, ErrUserNotFound):
		return httpx.MappedError{
			Status:  http.StatusNotFound,
			Code:    "USER_NOT_FOUND",
			Message: "user not found",
		}
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
	case errors.Is(err, ErrUserConflict):
		return httpx.MappedError{
			Status:  http.StatusConflict,
			Code:    "USER_CONFLICT",
			Message: "user conflicts with existing data",
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
