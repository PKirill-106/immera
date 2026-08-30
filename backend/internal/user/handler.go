package user

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"immera/internal/auth"
	"immera/internal/platform/httpx"
)

type userService interface {
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserSettings(ctx context.Context, id uuid.UUID) (UserSettings, error)
	UpdateUser(ctx context.Context, id uuid.UUID, user UpdateUserParams) error
	UpdateSettings(ctx context.Context, id uuid.UUID, settings UpdateSettingsParams) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
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

func (h *Handler) ProtectedRoutes(router chi.Router) {
	router.Get("/users/me", h.GetMe)
	router.Get("/users/me/settings", h.GetUserSettings)
	router.Put("/users/me", h.UpdateUser)
	router.Put("/users/me/settings", h.UpdateUserSettings)
	router.Delete("/users/me", h.DeleteUser)
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authenticatedUserID(w, r)
	if !ok {
		return
	}

	foundUser, err := h.service.GetByID(r.Context(), userID)
	if err != nil {
		h.writeMappedError(w, err, "failed to get user", "user_id", userID.String())
		return
	}

	response := toUserByIDResponse(foundUser)

	if err := httpx.WriteJSON(w, http.StatusOK, response); err != nil {
		h.log.Error(
			"failed to write user response",
			"user_id", userID.String(),
			"error", err,
		)
	}
}

func (h *Handler) GetUserSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authenticatedUserID(w, r)
	if !ok {
		return
	}

	foundSettings, err := h.service.GetUserSettings(r.Context(), userID)

	if err != nil {
		h.writeMappedError(w, err, "failed to get user settings", "user_id", userID.String())
		return
	}

	response := toUserSettingsResponse(foundSettings)

	if err := httpx.WriteJSON(w, http.StatusOK, response); err != nil {
		h.log.Error(
			"failed to write user settings response",
			"user_id", userID.String(),
			"error", err,
		)
	}
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authenticatedUserID(w, r)
	if !ok {
		return
	}

	var req updateUserRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode update user request", "user_id", userID.String())
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode update user request", "user_id", userID.String())
		return
	}

	err := h.service.UpdateUser(
		r.Context(),
		userID,
		UpdateUserParams{
			Name:        req.Name,
			Email:       req.Email,
			PhoneNumber: req.PhoneNumber,
		},
	)
	if err != nil {
		h.writeMappedError(w, err, "failed to update user", "user_id", userID.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateUserSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authenticatedUserID(w, r)
	if !ok {
		return
	}

	var req updateSettingsRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode update settings request", "user_id", userID.String())
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode update settings request", "user_id", userID.String())
		return
	}

	err := h.service.UpdateSettings(
		r.Context(),
		userID,
		UpdateSettingsParams{
			DefaultLanguage: req.DefaultLanguage,
			Theme:           req.Theme,
		},
	)
	if err != nil {
		h.writeMappedError(w, err, "failed to update settings", "user_id", userID.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.authenticatedUserID(w, r)
	if !ok {
		return
	}

	if err := h.service.DeleteUser(r.Context(), userID); err != nil {
		h.writeMappedError(w, err, "failed to delete user", "user_id", userID.String())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) authenticatedUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if ok {
		return userID, true
	}

	h.log.Error("authenticated user ID missing from context")
	if err := httpx.WriteError(
		w,
		http.StatusInternalServerError,
		"INTERNAL_ERROR",
		"internal server error",
	); err != nil {
		h.log.Error(
			"failed to write internal error response",
			"error", err,
		)
	}

	return uuid.Nil, false
}
