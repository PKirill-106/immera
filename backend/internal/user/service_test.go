package user

import (
	"errors"
	"strings"
	"testing"
)

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
