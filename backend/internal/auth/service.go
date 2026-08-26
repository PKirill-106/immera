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

	name, err := normalizeOptionalField(newUser.Name, 25, ErrNameTooLong)
	if err != nil {
		return err
	}

	phoneNumber, err := normalizeOptionalField(newUser.PhoneNumber, 15, ErrPhoneNumberTooLong)
	if err != nil {
		return err
	}

	if err := validatePassword(newUser.Password); err != nil {
		return err
	}

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email || utf8.RuneCountInString(email) > 50 {
		return ErrInvalidEmail
	}

	hash, err := hashPassword(newUser.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	err = s.repo.RegisterUser(ctx, RegisterUserParams{
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
