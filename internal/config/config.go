package config

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

var (
	ErrConfigNotFound     = errors.New("config file not found")
	ErrFailedToLoadEnv    = errors.New("failed to load .env file")
	ErrFailedToLoadConfig = errors.New("failed to load config")
)

type HTTPServer struct {
	Address      string        `yaml:"address" env:"HTTP_ADDRESS" env-default:":5050"`
	WriteTimeout time.Duration `yaml:"write-timeout" env-default:"5s"`
	ReadTimeout  time.Duration `yaml:"read-timeout" env-default:"5s"`
	IdleTimeout  time.Duration `yaml:"idle-timeout" env-default:"60s"`
}

type Config struct {
	AppEnv      string     `env:"APP_ENV" env-default:"local"`
	DatabaseURL string     `env:"DATABASE_URL" env-required:""`
	HTTPServer  HTTPServer `yaml:"server"`
}

func Load() (Config, error) {
	var cfg Config

	if err := godotenv.Load(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Config{}, ErrFailedToLoadEnv
		}
	}

	configPath := strings.TrimSpace(os.Getenv("CONFIG_PATH"))

	if configPath == "" {
		configPath = "./config/config.yml"
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, ErrConfigNotFound
		}

		return Config{}, ErrFailedToLoadConfig
	}

	return cfg, nil
}
