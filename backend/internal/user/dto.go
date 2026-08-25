package user

import "github.com/google/uuid"

type userByIDResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
}
type userSettingsResponse struct {
	ID              uuid.UUID `json:"id"`
	DefaultLanguage string    `json:"default_language"`
	Theme           string    `json:"theme"`
}
type updateUserRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}
