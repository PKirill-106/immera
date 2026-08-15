package user

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type userService interface {
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
}

type Handler struct {
	service userService
	log     *slog.Logger
}

func NewHandler(service userService, log *slog.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) Routes(router chi.Router) {
	router.Get("/users/{userID}", h.GetByID)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "userID"))

	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	foundUser, err := h.service.GetByID(r.Context(), id)

	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			http.Error(w, "user not found", http.StatusNotFound)
		default:
			h.log.Error(
				"failed to get user",
				"user_id", id.String(),
				"error", err,
			)

			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := touserByIDResponse(foundUser)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode user response",
			"user_id", id.String(),
			"error", err,
		)
	}
}
