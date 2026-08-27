package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo       Repository
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(
	repo Repository,
	jwtSecret []byte,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *Service {
	return &Service{
		repo:       repo,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
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
	if phoneNumber != nil {
		if err := validatePhoneNumber(*phoneNumber); err != nil {
			return err
		}
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

func (s *Service) Login(ctx context.Context, params LoginParams) (TokenPair, error) {
	email := strings.ToLower(strings.TrimSpace(params.Email))

	credentials, err := s.repo.GetCredentialsByEmail(ctx, email)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return TokenPair{}, ErrInvalidCredentials
		}

		return TokenPair{}, fmt.Errorf("login: get credentials: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(credentials.PasswordHash), []byte(params.Password))

	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return TokenPair{}, ErrInvalidCredentials
		}

		return TokenPair{}, fmt.Errorf("login: compare password: %w", err)
	}

	accessToken, err := generateAccessToken(
		credentials.ID,
		s.jwtSecret,
		s.accessTTL,
	)

	if err != nil {
		return TokenPair{}, fmt.Errorf("login: generate access token: %w", err)
	}

	refreshToken, err := generateRefreshToken()

	if err != nil {
		return TokenPair{}, fmt.Errorf("login: generate refresh token: %w", err)
	}

	refreshTokenHash := hashRefreshToken(refreshToken)
	refreshExpiresAt := time.Now().Add(s.refreshTTL)

	err = s.repo.CreateRefreshSession(
		ctx,
		credentials.ID,
		refreshTokenHash,
		refreshExpiresAt,
	)

	if err != nil {
		return TokenPair{}, fmt.Errorf("login: create refresh session: %w", err)
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
