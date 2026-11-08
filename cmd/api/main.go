package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/handler"
	"url-shortener/internal/lib/logger/sl"
	appmiddleware "url-shortener/internal/middleware"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	envProd  = "prod"
	envLocal = "local"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	logger := setupLogger(cfg.Env)

	if err != nil {
		logger.Error("failed to load configuration", sl.Err(err))
		os.Exit(1)
	}

	logger.Info(
		"starting service",
		slog.String("service_name", cfg.AppName),
		slog.String("env", cfg.Env),
		slog.String("address", cfg.HTTPServer.Address))

	logger.Debug("debug messages enabled")

	// init dbpool
	pool, err := initDbPool(ctx, cfg.DatabaseURL)

	if err != nil {
		logger.Error("database error", sl.Err(err))
		os.Exit(1)
	}

	r := chi.NewRouter()
	repo := repository.NewURLRepository(pool)
	s := service.NewURLService(cfg.HTTPServer.Address, cfg.AliasCodeLength, cfg.AliasGenMaxAttempts, repo)
	h := handler.NewURLHandler(s, logger)

	// routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
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

	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      r,
		WriteTimeout: cfg.HTTPServer.WriteTimeout,
		ReadTimeout:  cfg.HTTPServer.ReadTimeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	sigCtx, stop := signal.NotifyContext(
		ctx,
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go runServer(server, logger)

	<-sigCtx.Done()
	logger.Info("shutdown signal recieved")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server graceful shutdown error", "error", err)
	}

	logger.Info("HTTP server stopped")
}

func initDbPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		return nil, fmt.Errorf("failed to init db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return pool, nil
}

func setupLogger(env string) *slog.Logger {
	var opts *slog.HandlerOptions
	var handler slog.Handler

	switch env {
	case envLocal:
		opts = &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		handler = slog.NewTextHandler(os.Stdout, opts)
	case envProd:
		opts = &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
		handler = slog.NewJSONHandler(os.Stdout, opts)

	default:
		log.Printf("unknown app environment option: %s", env)
		opts = &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func runServer(s *http.Server, logger *slog.Logger) {
	logger.Info("starting HTTP server", slog.String("addr", s.Addr))
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("HTTP server error", "error", err)
		os.Exit(1)
	}
}
