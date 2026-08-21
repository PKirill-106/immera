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
