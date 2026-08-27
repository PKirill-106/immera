package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"immera/internal/platform/httpx"
)

type contextKey struct{}

var userIDContextKey contextKey

func Middleware(secret []byte, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeUnauthorized(w, log)
				return
			}

			tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || strings.TrimSpace(tokenString) == "" {
				writeUnauthorized(w, log)
				return
			}

			token, err := jwt.Parse(
				tokenString,
				func(token *jwt.Token) (any, error) {
					if token.Method != jwt.SigningMethodHS256 {
						return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
					}

					return secret, nil
				},
				jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
				jwt.WithExpirationRequired(),
			)

			if err != nil {
				writeUnauthorized(w, log)
				return
			}

			if !token.Valid {
				writeUnauthorized(w, log)
				return
			}

			sub, err := token.Claims.GetSubject()
			if err != nil {
				writeUnauthorized(w, log)
				return
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				writeUnauthorized(w, log)
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(
	w http.ResponseWriter,
	log *slog.Logger,
) {
	if err := httpx.WriteError(
		w,
		http.StatusUnauthorized,
		"UNAUTHORIZED",
		"unauthorized",
	); err != nil {
		log.Error(
			"failed to write unauthorized response",
			"error", err,
		)
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDContextKey).(uuid.UUID)
	return userID, ok
}
