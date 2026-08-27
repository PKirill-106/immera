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
	case errors.Is(err, ErrPasswordTooShort),
		errors.Is(err, ErrPasswordTooLong),
		errors.Is(err, ErrPasswordMissingNumber),
		errors.Is(err, ErrPasswordMissingSpecial):
		return httpx.MappedError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_PASSWORD",
			Message: "password must be 8 to 40 characters and contain a number and special character",
		}
	case errors.Is(err, ErrNameTooLong):
		return httpx.MappedError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_NAME",
			Message: "name must be at most 25 characters",
		}
	case errors.Is(err, ErrPhoneNumberTooLong):
		return httpx.MappedError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_PHONE_NUMBER",
			Message: "phone number must be at most 15 characters",
		}
	case errors.Is(err, ErrInvalidCredentials):
		return httpx.MappedError{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_CREDENTIALS",
			Message: "invalid credentials",
		}
	case errors.Is(err, ErrRefreshTokenNotFound):
		return httpx.MappedError{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_REFRESH_TOKEN",
			Message: "invalid refresh token",
		}
	case errors.Is(err, ErrRefreshTokenExpired):
		return httpx.MappedError{
			Status:  http.StatusUnauthorized,
			Code:    "REFRESH_TOKEN_EXPIRED",
			Message: "refresh token expired",
		}
	case errors.Is(err, ErrVerificationTokenNotFound):
		return httpx.MappedError{
			Status:  http.StatusNotFound,
			Code:    "VERIFICATION_TOKEN_NOT_FOUND",
			Message: "verification token not found",
		}
	case errors.Is(err, ErrVerificationTokenExpired):
		return httpx.MappedError{
			Status:  http.StatusBadRequest,
			Code:    "VERIFICATION_TOKEN_EXPIRED",
			Message: "verification token expired",
		}
	case errors.Is(err, ErrVerificationTokenUsed):
		return httpx.MappedError{
			Status:  http.StatusConflict,
			Code:    "VERIFICATION_TOKEN_USED",
			Message: "verification token already used",
		}
	case errors.Is(err, ErrEmailAlreadyVerified):
		return httpx.MappedError{
			Status:  http.StatusConflict,
			Code:    "EMAIL_ALREADY_VERIFIED",
			Message: "email already verified",
		}
	case errors.Is(err, ErrUserNotFound):
		return httpx.MappedError{
			Status:  http.StatusNotFound,
			Code:    "USER_NOT_FOUND",
			Message: "user not found",
		}
	case errors.Is(err, ErrInvalidEmail):
		return httpx.MappedError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_EMAIL",
			Message: "invalid email",
		}
	case errors.Is(err, ErrInvalidPhoneNumber):
		return httpx.MappedError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_PHONE_NUMBER",
			Message: "invalid phone number",
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
