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
	repo                 Repository
	emailSender          EmailSender
	jwtSecret            []byte
	accessTTL            time.Duration
	refreshTTL           time.Duration
	emailVerificationTTL time.Duration
}

func NewService(
	repo Repository,
	emailSender EmailSender,
	jwtSecret []byte,
	accessTTL time.Duration,
	refreshTTL time.Duration,
	emailVerificationTTL time.Duration,
) *Service {
	return &Service{
		repo:                 repo,
		emailSender:          emailSender,
		jwtSecret:            jwtSecret,
		accessTTL:            accessTTL,
		refreshTTL:           refreshTTL,
		emailVerificationTTL: emailVerificationTTL,
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

	verificationToken, err := generateVerificationToken()
	if err != nil {
		return fmt.Errorf("generate verification token: %w", err)
	}

	err = s.repo.Register(ctx, RegisterParams{
		Name:                       name,
		Email:                      parsedEmail.Address,
		PhoneNumber:                phoneNumber,
		PasswordHash:               hash,
		VerificationTokenHash:      hashVerificationToken(verificationToken),
		VerificationTokenExpiresAt: time.Now().Add(s.emailVerificationTTL),
	})
	if err != nil {
		return fmt.Errorf("register user: %w", err)
	}

	if err := s.emailSender.SendVerificationEmail(ctx, parsedEmail.Address, verificationToken); err != nil {
		return fmt.Errorf("send registration verification email: %w", err)
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

func (s *Service) Refresh(ctx context.Context, params RefreshParams) (TokenPair, error) {
	if strings.TrimSpace(params.RefreshToken) == "" {
		return TokenPair{}, ErrInvalidRequest
	}

	oldTokenHash := hashRefreshToken(params.RefreshToken)
	session, err := s.repo.GetRefreshSessionByTokenHash(ctx, oldTokenHash)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenNotFound) {
			return TokenPair{}, ErrRefreshTokenNotFound
		}

		return TokenPair{}, fmt.Errorf("refresh: get refresh session: %w", err)
	}

	if !time.Now().Before(session.ExpiresAt) {
		return TokenPair{}, ErrRefreshTokenExpired
	}

	accessToken, err := generateAccessToken(session.UserID, s.jwtSecret, s.accessTTL)
	if err != nil {
		return TokenPair{}, fmt.Errorf("refresh: generate access token: %w", err)
	}

	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("refresh: generate refresh token: %w", err)
	}

	newRefreshTokenHash := hashRefreshToken(newRefreshToken)
	err = s.repo.RotateRefreshSession(
		ctx,
		oldTokenHash,
		session.UserID,
		newRefreshTokenHash,
		time.Now().Add(s.refreshTTL),
	)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenNotFound) {
			return TokenPair{}, ErrRefreshTokenNotFound
		}

		return TokenPair{}, fmt.Errorf("refresh: rotate refresh session: %w", err)
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *Service) Logout(ctx context.Context, params LogoutParams) error {
	if strings.TrimSpace(params.RefreshToken) == "" {
		return ErrInvalidRequest
	}

	if err := s.repo.DeleteRefreshSessionByTokenHash(
		ctx,
		hashRefreshToken(params.RefreshToken),
	); err != nil {
		return fmt.Errorf("logout: delete refresh session: %w", err)
	}

	return nil
}
