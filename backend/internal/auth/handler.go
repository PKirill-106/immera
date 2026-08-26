package auth

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	// "github.com/google/uuid"
)

type authService interface {
	RegisterUser(ctx context.Context, newUser RegisterUserDTO) error
	// LoginUser(ctx context.Context, email string, passwordHash string) (string, error)
	// LogoutUser(ctx context.Context, id uuid.UUID) error
	// RefreshToken(ctx context.Context, token string) (string, error)
}

type Handler struct {
	service authService
	log     *slog.Logger
}

func NewHandler(service authService, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) Routes(router chi.Router) {
	router.Post("/auth/register", h.RegisterUser)
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserDTO
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode register user request")
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode register user request")
		return
	}

	err := h.service.RegisterUser(r.Context(), req)
	if err != nil {
		h.writeMappedError(w, err, "failed to register user")
		return
	}

	w.WriteHeader(http.StatusCreated)
}
