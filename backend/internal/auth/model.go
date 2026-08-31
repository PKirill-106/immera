package auth

import (
	"time"

	"github.com/google/uuid"
)

type RegisterParams struct {
	Name                       *string
	Email                      string
	PhoneNumber                *string
	PasswordHash               string
	VerificationTokenHash      string
	VerificationTokenExpiresAt time.Time
}

type LoginParams struct {
	Email    string
	Password string
}

type RefreshParams struct {
	RefreshToken string
}

type LogoutParams struct {
	RefreshToken string
}

type VerifyEmailParams struct {
	Token string
}

type ResendVerificationParams struct {
	Email string
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

type EmailVerification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	UsedAt    *time.Time
}

type UserVerificationStatus struct {
	ID              uuid.UUID
	Email           string
	EmailVerifiedAt *time.Time
}
