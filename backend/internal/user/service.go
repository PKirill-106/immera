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
		return User{}, ErrInvalidID
	}

	foundUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}

	return foundUser, nil
}
