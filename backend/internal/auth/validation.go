package auth

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func normalizeOptionalField(value *string, maxCharacters int, tooLongError error) (*string, error) {
	if value == nil {
		return nil, nil
	}

	normalized := strings.TrimSpace(*value)
	if utf8.RuneCountInString(normalized) > maxCharacters {
		return nil, tooLongError
	}

	return &normalized, nil
}

func validatePassword(password string) error {
	characterCount := utf8.RuneCountInString(password)
	if characterCount < 8 {
		return ErrPasswordTooShort
	}

	if characterCount > 40 || len(password) > 72 {
		return ErrPasswordTooLong
	}

	hasNumber := false
	hasSpecial := false

	for _, character := range password {
		if unicode.IsDigit(character) {
			hasNumber = true
		}
		if unicode.IsPunct(character) || unicode.IsSymbol(character) {
			hasSpecial = true
		}
	}

	if !hasNumber {
		return ErrPasswordMissingNumber
	}
	if !hasSpecial {
		return ErrPasswordMissingSpecial
	}

	return nil
}
