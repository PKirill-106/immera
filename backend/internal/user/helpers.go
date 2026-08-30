package user

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

func normalizeUpdateUser(user UpdateUserParams) (UpdateUserParams, error) {
	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.PhoneNumber = strings.TrimSpace(user.PhoneNumber)

	if user.Name == "" || utf8.RuneCountInString(user.Name) > 30 {
		return UpdateUserParams{}, ErrInvalidUserData
	}

	parsedEmail, err := mail.ParseAddress(user.Email)
	if err != nil || parsedEmail.Address != user.Email || utf8.RuneCountInString(user.Email) > 50 {
		return UpdateUserParams{}, ErrInvalidUserData
	}

	if user.PhoneNumber == "" || utf8.RuneCountInString(user.PhoneNumber) > 15 {
		return UpdateUserParams{}, ErrInvalidUserData
	}

	return user, nil
}
func normalizeUpdateSettings(settings UpdateSettingsParams) (UpdateSettingsParams, bool) {
	settings.DefaultLanguage = strings.ToLower(strings.TrimSpace(settings.DefaultLanguage))
	settings.Theme = strings.ToLower(strings.TrimSpace(settings.Theme))

	if settings.DefaultLanguage == "" ||
		utf8.RuneCountInString(settings.DefaultLanguage) > 10 ||
		settings.Theme == "" ||
		utf8.RuneCountInString(settings.Theme) > 10 {
		return UpdateSettingsParams{}, false
	}

	return settings, true
}
