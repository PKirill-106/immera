package user

import (
	"context"
	"errors"
	"immera/internal/platform/httpx"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type userService interface {
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserSettings(ctx context.Context, id uuid.UUID) (UserSettings, error)
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
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "userID"))

	if err != nil {
		if writeErr := httpx.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_USER_ID",
			"invalid user id",
		); writeErr != nil {
			h.log.Error(
				"failed to write error response",
				"error", writeErr,
			)
		}

		return
	}

	foundUser, err := h.service.GetByID(r.Context(), id)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidUserID):
			if writeErr := httpx.WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_USER_ID",
				"invalid user id",
			); writeErr != nil {
				h.log.Error(
					"failed to write error response",
					"user_id", id.String(),
					"error", writeErr,
				)
			}

		case errors.Is(err, ErrUserNotFound):
			if writeErr := httpx.WriteError(
				w,
				http.StatusNotFound,
				"USER_NOT_FOUND",
				"user not found",
			); writeErr != nil {
				h.log.Error(
					"failed to write error response",
					"user_id", id.String(),
					"error", writeErr,
				)
			}

		default:
			h.log.Error(
				"failed to get user",
				"user_id", id.String(),
				"error", err,
			)

			if writeErr := httpx.WriteError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"internal server error",
			); writeErr != nil {
				h.log.Error(
					"failed to write error response",
					"user_id", id.String(),
					"error", writeErr,
				)
			}
		}

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
		if writeErr := httpx.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_USER_ID",
			"invalid user id",
		); writeErr != nil {
			h.log.Error(
				"failed to write error response",
				"error", writeErr,
			)
		}

		return
	}

	foundSettings, err := h.service.GetUserSettings(r.Context(), id)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidUserID):
			if writeErr := httpx.WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_USER_ID",
				"invalid user id",
			); writeErr != nil {
				h.log.Error(
					"failed to write error response",
					"user_id", id.String(),
					"error", writeErr,
				)
			}
		case errors.Is(err, ErrUserNotFound):
			if writeErr := httpx.WriteError(
				w,
				http.StatusNotFound,
				"USER_NOT_FOUND",
				"user not found",
			); writeErr != nil {
				h.log.Error(
					"failed to write error response",
					"user_id", id.String(),
					"error", writeErr,
				)
			}

		default:
			h.log.Error(
				"failed to get user settings",
				"user_id", id.String(),
				"error", err,
			)

			if writeErr := httpx.WriteError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"internal server error",
			); writeErr != nil {
				h.log.Error(
					"failed to write error response",
					"user_id", id.String(),
					"error", writeErr,
				)
			}
		}

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
