package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type stubRepository struct {
	registered           RegisterParams
	registerCalled       bool
	registerErr          error
	credentials          UserCredentials
	credentialsEmail     string
	getCredentialsCalled bool
	getCredentialsErr    error
	refreshUserID        uuid.UUID
	refreshTokenHash     string
	refreshExpiresAt     time.Time
	refreshSessionCalled bool
	refreshSessionErr    error
	refreshSessions      map[string]RefreshSession
	refreshSession       RefreshSession
	getRefreshTokenHash  string
	getRefreshCalled     bool
	getRefreshErr        error
	rotateOldTokenHash   string
	rotateNewTokenHash   string
	rotateUserID         uuid.UUID
	rotateExpiresAt      time.Time
	rotateCalled         bool
	rotateErr            error
}

func (r *stubRepository) Register(_ context.Context, registration RegisterParams) error {
	r.registerCalled = true
	r.registered = registration
	return r.registerErr
}

func (r *stubRepository) GetCredentialsByEmail(_ context.Context, email string) (UserCredentials, error) {
	r.getCredentialsCalled = true
	r.credentialsEmail = email
	return r.credentials, r.getCredentialsErr
}

func (r *stubRepository) CreateRefreshSession(
	_ context.Context,
	userID uuid.UUID,
	refreshTokenHash string,
	refreshExpiresAt time.Time,
) error {
	r.refreshSessionCalled = true
	r.refreshUserID = userID
	r.refreshTokenHash = refreshTokenHash
	r.refreshExpiresAt = refreshExpiresAt
	return r.refreshSessionErr
}

func (r *stubRepository) GetRefreshSessionByTokenHash(
	_ context.Context,
	tokenHash string,
) (RefreshSession, error) {
	r.getRefreshCalled = true
	r.getRefreshTokenHash = tokenHash
	if r.getRefreshErr != nil {
		return RefreshSession{}, r.getRefreshErr
	}
	if r.refreshSessions != nil {
		session, ok := r.refreshSessions[tokenHash]
		if !ok {
			return RefreshSession{}, ErrRefreshTokenNotFound
		}
		return session, nil
	}

	return r.refreshSession, nil
}

func (r *stubRepository) RotateRefreshSession(
	_ context.Context,
	oldTokenHash string,
	userID uuid.UUID,
	newTokenHash string,
	newExpiresAt time.Time,
) error {
	r.rotateCalled = true
	r.rotateOldTokenHash = oldTokenHash
	r.rotateUserID = userID
	r.rotateNewTokenHash = newTokenHash
	r.rotateExpiresAt = newExpiresAt
	if r.rotateErr != nil {
		return r.rotateErr
	}
	if r.refreshSessions != nil {
		if _, ok := r.refreshSessions[oldTokenHash]; !ok {
			return ErrRefreshTokenNotFound
		}
		delete(r.refreshSessions, oldTokenHash)
		r.refreshSessions[newTokenHash] = RefreshSession{UserID: userID, ExpiresAt: newExpiresAt}
	}

	return nil
}

func TestRegisterNormalizesAndHashesInput(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{}
	service := newTestService(repository)

	err := service.Register(context.Background(), RegisterDTO{
		Name:        stringPointer("  Jane Doe  "),
		Email:       "  JANE@EXAMPLE.COM  ",
		PhoneNumber: stringPointer("  +48123456789  "),
		Password:    "password1!",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if !repository.registerCalled {
		t.Fatal("repository was not called")
	}
	if repository.registered.Name == nil || *repository.registered.Name != "Jane Doe" {
		t.Fatalf("registered name = %v, want Jane Doe", repository.registered.Name)
	}
	if repository.registered.Email != "jane@example.com" {
		t.Fatalf("registered email = %q, want jane@example.com", repository.registered.Email)
	}
	if repository.registered.PhoneNumber == nil || *repository.registered.PhoneNumber != "+48123456789" {
		t.Fatalf("registered phone number = %v, want +48123456789", repository.registered.PhoneNumber)
	}
	if repository.registered.PasswordHash == "password1!" {
		t.Fatal("repository received the raw password")
	}
	if err := comparePassword(repository.registered.PasswordHash, "password1!"); err != nil {
		t.Fatalf("stored password hash does not match password: %v", err)
	}
}

func TestRegisterAllowsOmittedOptionalFields(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{}
	service := newTestService(repository)

	err := service.Register(context.Background(), RegisterDTO{
		Email:    "jane@example.com",
		Password: "password1!",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repository.registered.Name != nil {
		t.Fatalf("registered name = %v, want nil", repository.registered.Name)
	}
	if repository.registered.PhoneNumber != nil {
		t.Fatalf("registered phone number = %v, want nil", repository.registered.PhoneNumber)
	}
}

func TestRegisterPreservesEmptyOptionalNameForLaterPolicyDecision(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{}
	service := newTestService(repository)

	err := service.Register(context.Background(), RegisterDTO{
		Name:     stringPointer("   "),
		Email:    "jane@example.com",
		Password: "password1!",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repository.registered.Name == nil || *repository.registered.Name != "" {
		t.Fatalf("registered name = %v, want non-nil empty string", repository.registered.Name)
	}
	if repository.registered.PhoneNumber != nil {
		t.Fatalf("registered phone number = %v, want nil", repository.registered.PhoneNumber)
	}
}

func TestRegisterRejectsInvalidFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		registration RegisterDTO
		wantError    error
	}{
		{
			name:         "short password",
			registration: RegisterDTO{Email: "jane@example.com", Password: "pass1!"},
			wantError:    ErrPasswordTooShort,
		},
		{
			name:         "password over 40 characters",
			registration: RegisterDTO{Email: "jane@example.com", Password: strings.Repeat("a", 39) + "1!"},
			wantError:    ErrPasswordTooLong,
		},
		{
			name:         "password over 72 bytes",
			registration: RegisterDTO{Email: "jane@example.com", Password: strings.Repeat("🙂", 19) + "1!"},
			wantError:    ErrPasswordTooLong,
		},
		{
			name:         "password without number",
			registration: RegisterDTO{Email: "jane@example.com", Password: "password!"},
			wantError:    ErrPasswordMissingNumber,
		},
		{
			name:         "password without special character",
			registration: RegisterDTO{Email: "jane@example.com", Password: "password1"},
			wantError:    ErrPasswordMissingSpecial,
		},
		{
			name:         "name over 25 characters",
			registration: RegisterDTO{Name: stringPointer(strings.Repeat("a", 26)), Email: "jane@example.com", Password: "password1!"},
			wantError:    ErrNameTooLong,
		},
		{
			name:         "phone number over 15 characters",
			registration: RegisterDTO{Email: "jane@example.com", PhoneNumber: stringPointer(strings.Repeat("1", 16)), Password: "password1!"},
			wantError:    ErrPhoneNumberTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repository := &stubRepository{}
			service := newTestService(repository)

			err := service.Register(context.Background(), tt.registration)
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("Register() error = %v, want %v", err, tt.wantError)
			}
			if repository.registerCalled {
				t.Fatal("repository was called for invalid input")
			}
		})
	}
}

func TestLoginNormalizesEmailAndCreatesRefreshSession(t *testing.T) {
	t.Parallel()

	passwordHash, err := hashPassword("password1!")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	userID := uuid.New()
	repository := &stubRepository{
		credentials: UserCredentials{
			ID:           userID,
			Email:        "jane@example.com",
			PasswordHash: passwordHash,
		},
	}
	service := newTestService(repository)
	beforeExpiry := time.Now().Add(testRefreshTTL)

	tokens, err := service.Login(context.Background(), LoginParams{
		Email:    "  JANE@EXAMPLE.COM  ",
		Password: "password1!",
	})
	afterExpiry := time.Now().Add(testRefreshTTL)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if repository.credentialsEmail != "jane@example.com" {
		t.Fatalf("credentials email = %q, want jane@example.com", repository.credentialsEmail)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("tokens = %#v, want non-empty access and refresh tokens", tokens)
	}
	if !repository.refreshSessionCalled {
		t.Fatal("refresh session repository method was not called")
	}
	if repository.refreshUserID != userID {
		t.Fatalf("refresh user ID = %s, want %s", repository.refreshUserID, userID)
	}
	if repository.refreshTokenHash != hashRefreshToken(tokens.RefreshToken) {
		t.Fatal("stored refresh token hash does not match the returned refresh token")
	}
	if repository.refreshTokenHash == tokens.RefreshToken {
		t.Fatal("repository received the raw refresh token")
	}
	if repository.refreshExpiresAt.Before(beforeExpiry) || repository.refreshExpiresAt.After(afterExpiry) {
		t.Fatalf(
			"refresh expiry = %v, want between %v and %v",
			repository.refreshExpiresAt,
			beforeExpiry,
			afterExpiry,
		)
	}
}

func TestLoginReturnsInvalidCredentials(t *testing.T) {
	t.Parallel()

	passwordHash, err := hashPassword("password1!")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	tests := []struct {
		name       string
		repository *stubRepository
		password   string
		wantLookup bool
	}{
		{
			name: "user not found",
			repository: &stubRepository{
				getCredentialsErr: ErrUserNotFound,
			},
			password:   "password1!",
			wantLookup: true,
		},
		{
			name: "incorrect password",
			repository: &stubRepository{
				credentials: UserCredentials{
					ID:           uuid.New(),
					Email:        "jane@example.com",
					PasswordHash: passwordHash,
				},
			},
			password:   "incorrect1!",
			wantLookup: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := newTestService(tt.repository)
			tokens, err := service.Login(context.Background(), LoginParams{
				Email:    "jane@example.com",
				Password: tt.password,
			})
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
			}
			if tokens != (TokenPair{}) {
				t.Fatalf("tokens = %#v, want empty token pair", tokens)
			}
			if tt.repository.getCredentialsCalled != tt.wantLookup {
				t.Fatalf("credentials lookup called = %t, want %t", tt.repository.getCredentialsCalled, tt.wantLookup)
			}
			if tt.repository.refreshSessionCalled {
				t.Fatal("refresh session was created for invalid credentials")
			}
		})
	}
}

func TestLoginReturnsRefreshSessionErrorWithoutTokens(t *testing.T) {
	t.Parallel()

	passwordHash, err := hashPassword("password1!")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repositoryError := errors.New("database unavailable")
	repository := &stubRepository{
		credentials: UserCredentials{
			ID:           uuid.New(),
			Email:        "jane@example.com",
			PasswordHash: passwordHash,
		},
		refreshSessionErr: repositoryError,
	}
	service := newTestService(repository)

	tokens, err := service.Login(context.Background(), LoginParams{
		Email:    "jane@example.com",
		Password: "password1!",
	})
	if !errors.Is(err, repositoryError) {
		t.Fatalf("Login() error = %v, want wrapped %v", err, repositoryError)
	}
	if tokens != (TokenPair{}) {
		t.Fatalf("tokens = %#v, want empty token pair", tokens)
	}
}

func TestRefreshRotatesTokenAndInvalidatesOldToken(t *testing.T) {
	t.Parallel()

	oldToken := "old-refresh-token"
	oldTokenHash := hashRefreshToken(oldToken)
	userID := uuid.New()
	repository := &stubRepository{
		refreshSessions: map[string]RefreshSession{
			oldTokenHash: {
				UserID:    userID,
				ExpiresAt: time.Now().Add(time.Hour),
			},
		},
	}
	service := newTestService(repository)

	tokens, err := service.Refresh(context.Background(), RefreshParams{RefreshToken: oldToken})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" || tokens.RefreshToken == oldToken {
		t.Fatalf("tokens = %#v, want new access and refresh tokens", tokens)
	}
	if _, exists := repository.refreshSessions[oldTokenHash]; exists {
		t.Fatal("old refresh token hash still exists after rotation")
	}
	newTokenHash := hashRefreshToken(tokens.RefreshToken)
	newSession, exists := repository.refreshSessions[newTokenHash]
	if !exists {
		t.Fatal("new refresh token hash was not stored")
	}
	if newSession.UserID != userID {
		t.Fatalf("new refresh session user ID = %s, want %s", newSession.UserID, userID)
	}

	secondTokens, err := service.Refresh(context.Background(), RefreshParams{RefreshToken: oldToken})
	if !errors.Is(err, ErrRefreshTokenNotFound) {
		t.Fatalf("second Refresh() error = %v, want %v", err, ErrRefreshTokenNotFound)
	}
	if secondTokens != (TokenPair{}) {
		t.Fatalf("second tokens = %#v, want empty token pair", secondTokens)
	}
}

func TestRefreshRejectsUnknownToken(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{refreshSessions: make(map[string]RefreshSession)}
	service := newTestService(repository)

	tokens, err := service.Refresh(
		context.Background(),
		RefreshParams{RefreshToken: "unknown-refresh-token"},
	)
	if !errors.Is(err, ErrRefreshTokenNotFound) {
		t.Fatalf("Refresh() error = %v, want %v", err, ErrRefreshTokenNotFound)
	}
	if tokens != (TokenPair{}) || repository.rotateCalled {
		t.Fatalf("tokens = %#v, rotate called = %t", tokens, repository.rotateCalled)
	}
}

func TestRefreshRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{
		refreshSession: RefreshSession{
			UserID:    uuid.New(),
			ExpiresAt: time.Now().Add(-time.Minute),
		},
	}
	service := newTestService(repository)

	tokens, err := service.Refresh(
		context.Background(),
		RefreshParams{RefreshToken: "expired-refresh-token"},
	)
	if !errors.Is(err, ErrRefreshTokenExpired) {
		t.Fatalf("Refresh() error = %v, want %v", err, ErrRefreshTokenExpired)
	}
	if tokens != (TokenPair{}) || repository.rotateCalled {
		t.Fatalf("tokens = %#v, rotate called = %t", tokens, repository.rotateCalled)
	}
}

func TestRefreshWrapsRepositoryFailure(t *testing.T) {
	t.Parallel()

	repositoryError := errors.New("database unavailable")
	repository := &stubRepository{getRefreshErr: repositoryError}
	service := newTestService(repository)

	_, err := service.Refresh(context.Background(), RefreshParams{RefreshToken: "refresh-token"})
	if !errors.Is(err, repositoryError) {
		t.Fatalf("Refresh() error = %v, want wrapped %v", err, repositoryError)
	}
}

const (
	testAccessTTL  = 15 * time.Minute
	testRefreshTTL = 24 * time.Hour
)

var testJWTSecret = []byte("test-secret-that-is-at-least-32-bytes")

func newTestService(repository Repository) *Service {
	return NewService(repository, testJWTSecret, testAccessTTL, testRefreshTTL)
}

func stringPointer(value string) *string {
	return &value
}
