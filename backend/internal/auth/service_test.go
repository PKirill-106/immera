package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type stubRepository struct {
	registered RegisterParams
	called     bool
	err        error
}

func (r *stubRepository) Register(_ context.Context, registration RegisterParams) error {
	r.called = true
	r.registered = registration
	return r.err
}

func TestRegisterNormalizesAndHashesInput(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{}
	service := NewService(repository)

	err := service.Register(context.Background(), RegisterDTO{
		Name:        stringPointer("  Jane Doe  "),
		Email:       "  JANE@EXAMPLE.COM  ",
		PhoneNumber: stringPointer("  +48123456789  "),
		Password:    "password1!",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
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

func TestRegisterAllowsOmittedOptionalFields(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{}
	service := NewService(repository)

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

func TestRegisterPreservesEmptyOptionalFieldsForLaterPolicyDecision(t *testing.T) {
	t.Parallel()

	repository := &stubRepository{}
	service := NewService(repository)

	err := service.Register(context.Background(), RegisterDTO{
		Name:        stringPointer("   "),
		Email:       "jane@example.com",
		PhoneNumber: stringPointer("   "),
		Password:    "password1!",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repository.registered.Name == nil || *repository.registered.Name != "" {
		t.Fatalf("registered name = %v, want non-nil empty string", repository.registered.Name)
	}
	if repository.registered.PhoneNumber == nil || *repository.registered.PhoneNumber != "" {
		t.Fatalf("registered phone number = %v, want non-nil empty string", repository.registered.PhoneNumber)
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
			service := NewService(repository)

			err := service.Register(context.Background(), tt.registration)
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("Register() error = %v, want %v", err, tt.wantError)
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
