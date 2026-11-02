package main

import (
	"log"
	"log/slog"
	"os"
	"url-shortener/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("app configuration error: %s", err)
	}

	logger := setupLogger(cfg.AppEnv)

	logger.Info(
		"starting app",
		slog.String("app-env", cfg.AppEnv),
	)

}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger
	opts := &slog.HandlerOptions{}
	var handler slog.Handler

	switch env {
	case envDev:
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	case envLocal:
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case envProd:
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		log.Printf("unknown app environment option: %s", env)
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger = slog.New(handler)

	return logger
}
