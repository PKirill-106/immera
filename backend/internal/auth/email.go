package auth

import (
	"context"
	"log/slog"
)

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, email, token string) error
}

type DevelopmentEmailSender struct {
	log *slog.Logger
}

func NewDevelopmentEmailSender(log *slog.Logger) *DevelopmentEmailSender {
	return &DevelopmentEmailSender{log: log}
}

func (s *DevelopmentEmailSender) SendVerificationEmail(
	_ context.Context,
	email string,
	token string,
) error {
	s.log.Info(
		"development email verification token",
		"email", email,
		"token", token,
	)

	return nil
}
