package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type stubRepository struct {
	registered RegisterUserParams
	called     bool
	err        error
}

func (r *stubRepository) RegisterUser(_ context.Context, newUser RegisterUserParams) error {
	r.called = true
	r.registered = newUser
	return r.err
}

func TestRegisterUserNormalizesAndHashesInput(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{}
	service := NewService(repository)

	err := service.RegisterUser(context.Background(), RegisterUserDTO{
		Name:        stringPointer("  Jane Doe  "),
		Email:       "  JANE@EXAMPLE.COM  ",
		PhoneNumber: stringPointer("  +48123456789  "),
		Password:    "password1!",
	})
	if err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}
	if !repository.called {
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

func TestRegisterUserAllowsOmittedOptionalFields(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{}
	service := NewService(repository)

	err := service.RegisterUser(context.Background(), RegisterUserDTO{
		Email:    "jane@example.com",
		Password: "password1!",
	})
	if err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}
	if repository.registered.Name != nil {
		t.Fatalf("registered name = %v, want nil", repository.registered.Name)
	}
	if repository.registered.PhoneNumber != nil {
		t.Fatalf("registered phone number = %v, want nil", repository.registered.PhoneNumber)
	}
}

func TestRegisterUserPreservesEmptyOptionalFieldsForLaterPolicyDecision(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{}
	service := NewService(repository)

	err := service.RegisterUser(context.Background(), RegisterUserDTO{
		Name:        stringPointer("   "),
		Email:       "jane@example.com",
		PhoneNumber: stringPointer("   "),
		Password:    "password1!",
	})
	if err != nil {
		t.Fatalf("RegisterUser() error = %v", err)
	}
	if repository.registered.Name == nil || *repository.registered.Name != "" {
		t.Fatalf("registered name = %v, want non-nil empty string", repository.registered.Name)
	}
	if repository.registered.PhoneNumber == nil || *repository.registered.PhoneNumber != "" {
		t.Fatalf("registered phone number = %v, want non-nil empty string", repository.registered.PhoneNumber)
	}
}

func TestRegisterUserRejectsInvalidFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		newUser   RegisterUserDTO
		wantError error
	}{
		{
			name:      "short password",
			newUser:   RegisterUserDTO{Email: "jane@example.com", Password: "pass1!"},
			wantError: ErrPasswordTooShort,
		},
		{
			name:      "password over 40 characters",
			newUser:   RegisterUserDTO{Email: "jane@example.com", Password: strings.Repeat("a", 39) + "1!"},
			wantError: ErrPasswordTooLong,
		},
		{
			name:      "password over 72 bytes",
			newUser:   RegisterUserDTO{Email: "jane@example.com", Password: strings.Repeat("🙂", 19) + "1!"},
			wantError: ErrPasswordTooLong,
		},
		{
			name:      "password without number",
			newUser:   RegisterUserDTO{Email: "jane@example.com", Password: "password!"},
			wantError: ErrPasswordMissingNumber,
		},
		{
			name:      "password without special character",
			newUser:   RegisterUserDTO{Email: "jane@example.com", Password: "password1"},
			wantError: ErrPasswordMissingSpecial,
		},
		{
			name:      "name over 25 characters",
			newUser:   RegisterUserDTO{Name: stringPointer(strings.Repeat("a", 26)), Email: "jane@example.com", Password: "password1!"},
			wantError: ErrNameTooLong,
		},
		{
			name:      "phone number over 15 characters",
			newUser:   RegisterUserDTO{Email: "jane@example.com", PhoneNumber: stringPointer(strings.Repeat("1", 16)), Password: "password1!"},
			wantError: ErrPhoneNumberTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repository := &stubRepository{}
			service := NewService(repository)

			err := service.RegisterUser(context.Background(), tt.newUser)
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("RegisterUser() error = %v, want %v", err, tt.wantError)
			}
			if repository.called {
				t.Fatal("repository was called for invalid input")
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}
