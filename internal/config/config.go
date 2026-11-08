package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env                 string     `yaml:"app_env" env:"APP_ENV" env-default:"prod"`
	AppName             string     `yaml:"service_name" env:"APP_NAME" env-default:"url-shortener"`
	AliasGenMaxAttempts int        `yaml:"alias_generation_max_attempts" env:"ALIAS_GEN_MAX_ATTEMPTS" env-default:"5"`
	AliasCodeLength     int        `yaml:"alias_code_length" env:"ALIAS_CODE_LENGTH" env-default:"6"`
	DatabaseURL         string     `env:"DATABASE_URL" env-required:"true"`
	HTTPServer          HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Address      string        `yaml:"address" env:"HTTP_ADDRESS" env-default:":5050"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"5s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" env-default:"10s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	User         string        `yaml:"user" env-required:"true" env:"BASIC_AUTH_USER"`
	Password     string        `yaml:"password" env-required:"true" env:"BASIC_AUTH_PASSWORD"`
}

func Load() (Config, error) {
	var cfg Config

	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("failed to load env variables: %w", err)
	}

	configPath := strings.TrimSpace(os.Getenv("CONFIG_PATH"))

	if configPath == "" {
		configPath = "config/config.yml"
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	return cfg, nil
}
