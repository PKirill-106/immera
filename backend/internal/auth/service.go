package auth

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) RegisterUser(ctx context.Context, newUser RegisterUserDTO) error {
	email := strings.ToLower(strings.TrimSpace(newUser.Email))

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email || utf8.RuneCountInString(email) > 50 {
		return ErrInvalidEmail
	}

	hash, err := hashPassword(newUser.Password)
	if err != nil {
		return ErrInvalidCredentials
	}

	err = s.repo.RegisterUser(ctx, RegisterUserParams{
		newUser.Name,
		parsedEmail.Address,
		newUser.PhoneNumber,
		hash,
	})
	if err != nil {
		return fmt.Errorf("register user: %w", err)
	}

	return nil
}
