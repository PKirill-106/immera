package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type PingFunc func(context.Context) error

type Handler struct{ ping PingFunc }

type response struct {
	Status string `json:"status"`
}

func NewHandler(ping PingFunc) *Handler { return &Handler{ping: ping} }

func (h *Handler) Routes(router chi.Router) {
	router.Get("/health/live", h.Live)
	router.Get("/health/ready", h.Ready)
}

func (h *Handler) Live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, response{Status: "ok"})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, response{Status: "unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, response{Status: "ok"})
}

func writeJSON(w http.ResponseWriter, status int, value response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
