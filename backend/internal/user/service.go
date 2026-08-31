package user

import (
	"context"
	"fmt"

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
func (s *Service) UpdateSettings(
	ctx context.Context,
	id uuid.UUID,
	settings UpdateSettingsParams,
) error {
	if id == uuid.Nil {
		return ErrInvalidUserID
	}

	normalizedSettings, ok := normalizeUpdateSettings(settings)

	if !ok {
		return ErrInvalidSettingsData
	}

	err := s.repo.UpdateSettings(ctx, id, normalizedSettings)
	if err != nil {
		return fmt.Errorf("update settings: %w", err)
	}

	return nil
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidUserID
	}

	if err := s.repo.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return nil
}
