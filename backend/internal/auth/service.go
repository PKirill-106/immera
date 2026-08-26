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

func (s *Service) Register(ctx context.Context, registration RegisterDTO) error {
	email := strings.ToLower(strings.TrimSpace(registration.Email))

	name, err := normalizeOptionalField(registration.Name, 25, ErrNameTooLong)
	if err != nil {
		return err
	}

	phoneNumber, err := normalizeOptionalField(registration.PhoneNumber, 15, ErrPhoneNumberTooLong)
	if err != nil {
		return err
	}

	if err := validatePassword(registration.Password); err != nil {
		return err
	}

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email || utf8.RuneCountInString(email) > 50 {
		return ErrInvalidEmail
	}

	hash, err := hashPassword(registration.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	err = s.repo.Register(ctx, RegisterParams{
		Name:         name,
		Email:        parsedEmail.Address,
		PhoneNumber:  phoneNumber,
		PasswordHash: hash,
	})
	if err != nil {
		return fmt.Errorf("register user: %w", err)
	}

	return nil
}
