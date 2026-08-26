package auth

import "errors"

var (
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrPhoneNumberAlreadyExists = errors.New("phone number already exists")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrInvalidEmail             = errors.New("invalid email received")
	ErrInvalidRequest           = errors.New("invalid data received")
	ErrUnauthorized             = errors.New("unauthorized")
	ErrInvalidToken             = errors.New("invalid token")
	ErrExpiredToken             = errors.New("expired token")
)
