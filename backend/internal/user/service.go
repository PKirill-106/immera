package user

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (User, error) {
	if id == uuid.Nil {
		return User{}, ErrInvalidUserID
	}

	foundUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}

	return foundUser, nil
}

func (s *Service) GetUserSettings(ctx context.Context, id uuid.UUID) (UserSettings, error) {
	if id == uuid.Nil {
		return UserSettings{}, ErrInvalidUserID
	}

	foundSettings, err := s.repo.GetUserSettings(ctx, id)
	if err != nil {
		return UserSettings{}, fmt.Errorf("get user settings: %w", err)
	}

	return foundSettings, nil
}
func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, user UpdateUserParams) error {
	if id == uuid.Nil {
		return ErrInvalidUserID
	}

	normalizedUser, err := normalizeUpdateUser(user)
	if err != nil {
		return err
	}

	err = s.repo.UpdateUser(ctx, id, normalizedUser)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

func normalizeUpdateUser(user UpdateUserParams) (UpdateUserParams, error) {
	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.PhoneNumber = strings.TrimSpace(user.PhoneNumber)

	if user.Name == "" || utf8.RuneCountInString(user.Name) > 30 {
		return UpdateUserParams{}, ErrInvalidUserData
	}

	parsedEmail, err := mail.ParseAddress(user.Email)
	if err != nil || parsedEmail.Address != user.Email || utf8.RuneCountInString(user.Email) > 50 {
		return UpdateUserParams{}, ErrInvalidUserData
	}

	if user.PhoneNumber == "" || utf8.RuneCountInString(user.PhoneNumber) > 15 {
		return UpdateUserParams{}, ErrInvalidUserData
	}

	return user, nil
}
