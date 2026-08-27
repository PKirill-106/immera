package auth

import "errors"

var (
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrPhoneNumberAlreadyExists = errors.New("phone number already exists")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrInvalidEmail             = errors.New("invalid email received")
	ErrInvalidPhoneNumber       = errors.New("invalid phone number received")
	ErrInvalidRequest           = errors.New("invalid data received")

	ErrPasswordTooShort       = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong        = errors.New("password must be at most 40 characters and 72 bytes")
	ErrPasswordMissingNumber  = errors.New("password must contain a number")
	ErrPasswordMissingSpecial = errors.New("password must contain a special character")
	ErrNameTooLong            = errors.New("name must be at most 25 characters")
	ErrPhoneNumberTooLong     = errors.New("phone number must be at most 15 characters")

	ErrUnauthorized = errors.New("unauthorized")
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")

	ErrRefreshTokenNotFound      = errors.New("refresh token not found")
	ErrRefreshTokenExpired       = errors.New("refresh token expired")
	ErrVerificationTokenNotFound = errors.New("verification token not found")
	ErrVerificationTokenExpired  = errors.New("verification token expired")
	ErrVerificationTokenUsed     = errors.New("verification token already used")
	ErrEmailAlreadyVerified      = errors.New("email already verified")
)
