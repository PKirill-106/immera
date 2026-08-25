package user

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"immera/internal/platform/httpx"
)

type userService interface {
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserSettings(ctx context.Context, id uuid.UUID) (UserSettings, error)
	UpdateUser(ctx context.Context, id uuid.UUID, user UpdateUserParams) error
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
	router.Get("/users/{userID}/settings", h.GetUserSettings)
	router.Put("/users/{userID}", h.UpdateUser)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "userID"))

	if err != nil {
		h.writeMappedError(w, ErrInvalidUserID, "invalid user ID")
		return
	}

	foundUser, err := h.service.GetByID(r.Context(), id)

	if err != nil {
		h.writeMappedError(w, err, "failed to get user", "user_id", id.String())
		return
	}

	response := toUserByIDResponse(foundUser)

	if err := httpx.WriteJSON(w, http.StatusOK, response); err != nil {
		h.log.Error(
			"failed to write user response",
			"user_id", id.String(),
			"error", err,
		)
	}
}

func (h *Handler) GetUserSettings(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "userID"))

	if err != nil {
		h.writeMappedError(w, ErrInvalidUserID, "invalid user ID")
		return
	}

	foundSettings, err := h.service.GetUserSettings(r.Context(), id)

	if err != nil {
		h.writeMappedError(w, err, "failed to get user settings", "user_id", id.String())
		return
	}

	response := toUserSettingsResponse(foundSettings)

	if err := httpx.WriteJSON(w, http.StatusOK, response); err != nil {
		h.log.Error(
			"failed to write user settings response",
			"user_id", id.String(),
			"error", err,
		)
	}
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "userID"))

	if err != nil {
		h.writeMappedError(w, ErrInvalidUserID, "invalid user ID")
		return
	}

	var req updateUserRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode update user request", "user_id", id.String())
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode update user request", "user_id", id.String())
		return
	}

	err = h.service.UpdateUser(
		r.Context(),
		id,
		UpdateUserParams{
			Name:        req.Name,
			Email:       req.Email,
			PhoneNumber: req.PhoneNumber,
		},
	)
	if err != nil {
		h.writeMappedError(w, err, "failed to update user", "user_id", id.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
