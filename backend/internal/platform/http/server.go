package http

import (
	"log/slog"
	stdhttp "net/http"

	"immera/internal/platform/config"
)

func NewServer(cfg config.HTTP, handler stdhttp.Handler, log *slog.Logger) *stdhttp.Server {
	return &stdhttp.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
	}
}
