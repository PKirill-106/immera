package user

import "errors"

var (
	ErrInvalidRequest           = errors.New("invalid request")
	ErrInvalidUserData          = errors.New("invalid user data")
	ErrInvalidSettingsData      = errors.New("invalid user settings data")
	ErrInvalidUserID            = errors.New("invalid user id")
	ErrUserNotFound             = errors.New("user not found")
	ErrUserConflict             = errors.New("user conflicts with existing data")
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrPhoneNumberAlreadyExists = errors.New("phone number already exists")
)
