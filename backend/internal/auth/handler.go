package auth

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"immera/internal/platform/httpx"

	"github.com/go-chi/chi/v5"
)

type authService interface {
	Register(ctx context.Context, registration RegisterDTO) error
	Login(ctx context.Context, params LoginParams) (TokenPair, error)
	Refresh(ctx context.Context, params RefreshParams) (TokenPair, error)
	Logout(ctx context.Context, params LogoutParams) error
	VerifyEmail(ctx context.Context, params VerifyEmailParams) error
	ResendVerification(ctx context.Context, params ResendVerificationParams) error
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
	router.Post("/auth/register", h.Register)
	router.Post("/auth/login", h.Login)
	router.Post("/auth/refresh", h.Refresh)
	router.Post("/auth/logout", h.Logout)
	router.Post("/auth/verify-email", h.VerifyEmail)
	router.Post("/auth/resend-verification", h.ResendVerification)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterDTO
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

	err := h.service.Register(r.Context(), req)
	if err != nil {
		h.writeMappedError(w, err, "failed to register user")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequestDTO

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode login request")
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode login request")
		return
	}

	tokens, err := h.service.Login(
		r.Context(),
		LoginParams{
			Email:    req.Email,
			Password: req.Password,
		},
	)
	if err != nil {
		h.writeMappedError(w, err, "failed to login user")
		return
	}

	response := tokenPairResponseDTO{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	if err := httpx.WriteJSON(w, http.StatusOK, response); err != nil {
		h.log.Error(
			"failed to write login response",
			"error", err,
		)
	}
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequestDTO
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode refresh request")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode refresh request")
		return
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		h.writeMappedError(w, ErrInvalidRequest, "empty refresh token")
		return
	}

	tokens, err := h.service.Refresh(r.Context(), RefreshParams{RefreshToken: req.RefreshToken})
	if err != nil {
		h.writeMappedError(w, err, "failed to refresh tokens")
		return
	}

	if err := httpx.WriteJSON(w, http.StatusOK, tokenPairResponseDTO{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}); err != nil {
		h.log.Error("failed to write refresh response", "error", err)
	}
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequestDTO
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode logout request")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode logout request")
		return
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		h.writeMappedError(w, ErrInvalidRequest, "empty logout refresh token")
		return
	}

	if err := h.service.Logout(r.Context(), LogoutParams{RefreshToken: req.RefreshToken}); err != nil {
		h.writeMappedError(w, err, "failed to logout user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequestDTO
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode verify email request")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode verify email request")
		return
	}
	if strings.TrimSpace(req.Token) == "" {
		h.writeMappedError(w, ErrInvalidRequest, "empty email verification token")
		return
	}

	if err := h.service.VerifyEmail(r.Context(), VerifyEmailParams{Token: req.Token}); err != nil {
		h.writeMappedError(w, err, "failed to verify email")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req resendVerificationRequestDTO
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode resend verification request")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		h.writeMappedError(w, ErrInvalidRequest, "failed to decode resend verification request")
		return
	}

	if err := h.service.ResendVerification(
		r.Context(),
		ResendVerificationParams{Email: req.Email},
	); err != nil {
		h.writeMappedError(w, err, "failed to resend verification email")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
