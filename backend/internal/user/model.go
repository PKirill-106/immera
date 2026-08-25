package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID
	Name        string
	Email       string
	PhoneNumber string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserSettings struct {
	ID              uuid.UUID
	DefaultLanguage string
	Theme           string
	UserID          uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UpdateUserParams struct {
	Name        string
	Email       string
	PhoneNumber string
}
