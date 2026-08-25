package user

import (
	"errors"
	"net/http"

	"immera/internal/platform/httpx"
)

type mappedError struct {
	status  int
	code    string
	message string
}

func mapError(err error) mappedError {
	switch {
	case errors.Is(err, ErrInvalidUserID):
		return mappedError{
			status:  http.StatusBadRequest,
			code:    "INVALID_USER_ID",
			message: "invalid user id",
		}
	case errors.Is(err, ErrInvalidRequest):
		return mappedError{
			status:  http.StatusBadRequest,
			code:    "INVALID_REQUEST",
			message: "invalid request",
		}
	case errors.Is(err, ErrInvalidUserData):
		return mappedError{
			status:  http.StatusBadRequest,
			code:    "INVALID_USER_DATA",
			message: "invalid user data",
		}
	case errors.Is(err, ErrUserNotFound):
		return mappedError{
			status:  http.StatusNotFound,
			code:    "USER_NOT_FOUND",
			message: "user not found",
		}
	case errors.Is(err, ErrEmailAlreadyExists):
		return mappedError{
			status:  http.StatusConflict,
			code:    "EMAIL_ALREADY_EXISTS",
			message: "email already exists",
		}
	case errors.Is(err, ErrPhoneNumberAlreadyExists):
		return mappedError{
			status:  http.StatusConflict,
			code:    "PHONE_NUMBER_ALREADY_EXISTS",
			message: "phone number already exists",
		}
	case errors.Is(err, ErrUserConflict):
		return mappedError{
			status:  http.StatusConflict,
			code:    "USER_CONFLICT",
			message: "user conflicts with existing data",
		}
	default:
		return mappedError{
			status:  http.StatusInternalServerError,
			code:    "INTERNAL_ERROR",
			message: "internal server error",
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

	if mapped.status == http.StatusInternalServerError {
		attrs := appendLogError(logAttrs, err)
		h.log.Error(logMessage, attrs...)
	}

	if writeErr := httpx.WriteError(w, mapped.status, mapped.code, mapped.message); writeErr != nil {
		attrs := appendLogError(logAttrs, writeErr)
		h.log.Error("failed to write error response", attrs...)
	}
}

func appendLogError(attrs []any, err error) []any {
	result := make([]any, 0, len(attrs)+2)
	result = append(result, attrs...)
	return append(result, "error", err)
}
