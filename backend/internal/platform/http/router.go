package http

import (
	"log/slog"
	stdhttp "net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	platformmiddleware "immera/internal/platform/middleware"
)

type RouteRegistrar func(chi.Router)

func NewRouter(
	log *slog.Logger,
	allowedOrigins []string,
	infrastructureRoutes []RouteRegistrar,
	apiRoutes []RouteRegistrar,
) stdhttp.Handler {
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(platformmiddleware.RequestLogger(log))
	router.Use(chimiddleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "QUERY"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	for _, register := range infrastructureRoutes {
		register(router)
	}

	router.Route("/api/v1", func(api chi.Router) {
		for _, register := range apiRoutes {
			register(api)
		}
	})

	return router
}
