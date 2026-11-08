package router

import (
	"log/slog"
	"net/http"

	"url-shortener/internal/config"
	"url-shortener/internal/handler"
	appmiddleware "url-shortener/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	cfg config.Config,
	h *handler.URLHandler,
	logger *slog.Logger,
) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(
			middleware.RequestID,
			appmiddleware.LoggingMiddleware(logger),
			middleware.BasicAuth(cfg.AppName, map[string]string{
				cfg.HTTPServer.User: cfg.HTTPServer.Password,
			}),
			middleware.Recoverer,
		)

		r.Route("/urls", func(r chi.Router) {
			r.Post("/", h.Save)
			r.Get("/{code}", h.Get)
			r.Delete("/{code}", h.Delete)
			r.Patch("/{code}", h.Update)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(
			middleware.RequestID,
			appmiddleware.LoggingMiddleware(logger),
			middleware.Recoverer,
		)

		r.Get("/{code}", h.Redirect)
	})

	return r
}
