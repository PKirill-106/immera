package apidocs

import (
	"net/http"

	swaggerui "github.com/alexliesenfeld/go-swagger-ui"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	ui       http.Handler
	specPath string
}

func NewHandler(specURL string, specPath string) *Handler {
	ui := swaggerui.NewHandler(
		swaggerui.WithHTMLTitle("Immera API"),
		swaggerui.WithSpecURL(specURL),
		swaggerui.WithTryItOutEnabled(true),
		swaggerui.WithDocExpansion(swaggerui.DocExpansionList),
	)

	return &Handler{
		ui:       ui,
		specPath: specPath,
	}
}

func (h *Handler) Routes(router chi.Router) {
	router.Get("/openapi.yaml", h.Spec)
	router.Mount("/docs", h.ui)
	router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusPermanentRedirect)
	})
}

func (h *Handler) Spec(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.specPath)
}
