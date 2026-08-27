package auth

import (
	"time"

	"github.com/google/uuid"
)

type RegisterParams struct {
	Name         *string
	Email        string
	PhoneNumber  *string
	PasswordHash string
}
type LoginParams struct {
	Email    string
	Password string
}

type RefreshParams struct {
	RefreshToken string
}
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
type UserCredentials struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
}

type RefreshSession struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
}
