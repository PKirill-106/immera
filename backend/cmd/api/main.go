package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"immera/internal/health"
	"immera/internal/platform/config"
	"immera/internal/platform/database"
	httpserver "immera/internal/platform/http"
	"immera/internal/platform/logger"
	"immera/internal/user"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	log := logger.New(cfg.Environment)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer pool.Close()

	healthHandler := health.NewHandler(pool.Ping)

	userRepository := user.NewPostgresRepository(pool)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService, log)

	router := httpserver.NewRouter(log, cfg.HTTP.AllowedOrigins, []httpserver.RouteRegistrar{
		healthHandler.Routes,
	},
		[]httpserver.RouteRegistrar{
			userHandler.Routes,
		})
	server := httpserver.NewServer(cfg.HTTP, router, log)

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("HTTP server starting", "address", cfg.HTTP.Address, "environment", cfg.Environment)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	log.Info("HTTP server stopped")
	return nil
}
