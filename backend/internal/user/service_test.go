package user

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type stubUserRepository struct {
	deleteUserID uuid.UUID
	deleteCalled bool
	deleteErr    error
}

func (*stubUserRepository) GetByID(context.Context, uuid.UUID) (User, error) {
	return User{}, nil
}

func (*stubUserRepository) GetUserSettings(context.Context, uuid.UUID) (UserSettings, error) {
	return UserSettings{}, nil
}

func (*stubUserRepository) UpdateUser(context.Context, uuid.UUID, UpdateUserParams) error {
	return nil
}

func (*stubUserRepository) UpdateSettings(context.Context, uuid.UUID, UpdateSettingsParams) error {
	return nil
}

func (r *stubUserRepository) DeleteUser(_ context.Context, userID uuid.UUID) error {
	r.deleteCalled = true
	r.deleteUserID = userID
	return r.deleteErr
}

func TestNormalizeUpdateUser(t *testing.T) {
	t.Parallel()

	t.Run("normalizes valid input", func(t *testing.T) {
		t.Parallel()

		got, err := normalizeUpdateUser(UpdateUserParams{
			Name:        "  Jane Doe  ",
			Email:       "  JANE@EXAMPLE.COM  ",
			PhoneNumber: "  +48123456789  ",
		})
		if err != nil {
			t.Fatalf("normalizeUpdateUser() error = %v", err)
		}

		if got.Name != "Jane Doe" || got.Email != "jane@example.com" || got.PhoneNumber != "+48123456789" {
			t.Fatalf("normalizeUpdateUser() = %#v", got)
		}
	})

	tests := []struct {
		name string
		user UpdateUserParams
	}{
		{name: "empty name", user: UpdateUserParams{Email: "jane@example.com", PhoneNumber: "+48123456789"}},
		{name: "long name", user: UpdateUserParams{Name: strings.Repeat("a", 31), Email: "jane@example.com", PhoneNumber: "+48123456789"}},
		{name: "invalid email", user: UpdateUserParams{Name: "Jane", Email: "not-an-email", PhoneNumber: "+48123456789"}},
		{name: "long email", user: UpdateUserParams{Name: "Jane", Email: strings.Repeat("a", 39) + "@example.com", PhoneNumber: "+48123456789"}},
		{name: "empty phone", user: UpdateUserParams{Name: "Jane", Email: "jane@example.com"}},
		{name: "long phone", user: UpdateUserParams{Name: "Jane", Email: "jane@example.com", PhoneNumber: strings.Repeat("1", 16)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := normalizeUpdateUser(tt.user)
			if !errors.Is(err, ErrInvalidUserData) {
				t.Fatalf("normalizeUpdateUser() error = %v, want ErrInvalidUserData", err)
			}
		})
	}
}

func TestNormalizeUpdateSettings(t *testing.T) {
	t.Parallel()

	got, ok := normalizeUpdateSettings(UpdateSettingsParams{
		DefaultLanguage: "  EN  ",
		Theme:           "  DARK  ",
	})
	if !ok {
		t.Fatal("normalizeUpdateSettings() rejected valid settings")
	}
	if got.DefaultLanguage != "en" || got.Theme != "dark" {
		t.Fatalf("normalizeUpdateSettings() = %#v", got)
	}

	tests := []struct {
		name     string
		settings UpdateSettingsParams
	}{
		{name: "empty language", settings: UpdateSettingsParams{Theme: "dark"}},
		{name: "long language", settings: UpdateSettingsParams{DefaultLanguage: strings.Repeat("a", 11), Theme: "dark"}},
		{name: "empty theme", settings: UpdateSettingsParams{DefaultLanguage: "en"}},
		{name: "long theme", settings: UpdateSettingsParams{DefaultLanguage: "en", Theme: strings.Repeat("a", 11)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, ok := normalizeUpdateSettings(tt.settings); ok {
				t.Fatalf("normalizeUpdateSettings(%#v) accepted invalid settings", tt.settings)
			}
		})
	}
}

func TestDeleteUserDelegatesToRepository(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	repository := &stubUserRepository{}
	service := NewService(repository)

	if err := service.DeleteUser(context.Background(), userID); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
	if !repository.deleteCalled || repository.deleteUserID != userID {
		t.Fatalf(
			"repository DeleteUser called = %t with %s, want %s",
			repository.deleteCalled,
			repository.deleteUserID,
			userID,
		)
	}
}

func TestDeleteUserRejectsNilUserID(t *testing.T) {
	t.Parallel()

	repository := &stubUserRepository{}
	service := NewService(repository)

	err := service.DeleteUser(context.Background(), uuid.Nil)
	if !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("DeleteUser() error = %v, want %v", err, ErrInvalidUserID)
	}
	if repository.deleteCalled {
		t.Fatal("repository was called with a nil user ID")
	}
}

func TestDeleteUserPreservesRepositoryErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
	}{
		{name: "user not found", err: ErrUserNotFound},
		{name: "database failure", err: errors.New("database unavailable")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repository := &stubUserRepository{deleteErr: tt.err}
			service := NewService(repository)

			err := service.DeleteUser(context.Background(), uuid.New())
			if !errors.Is(err, tt.err) {
				t.Fatalf("DeleteUser() error = %v, want wrapped %v", err, tt.err)
			}
		})
	}
}
