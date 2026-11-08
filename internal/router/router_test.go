package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"url-shortener/internal/config"
	"url-shortener/internal/handler"
	"url-shortener/internal/lib/logger/fake"
	"url-shortener/internal/mocks"
	"url-shortener/internal/service"

	"github.com/joho/godotenv"
)

func TestBasicAuth_Unauthorized(t *testing.T) {
	logger := fake.Logger

	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf(
			"env load failed: %v",
			err,
		)
	}

	cfg := config.Config{
		AppName: "url-shortener",
		HTTPServer: config.HTTPServer{
			User:     os.Getenv("BASIC_AUTH_USER"),
			Password: os.Getenv("BASIC_AUTH_PASSWORD"),
		},
	}
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "Save",
			method: http.MethodPost,
			path:   "/api/urls/",
		},
		{
			name:   "Get",
			method: http.MethodGet,
			path:   "/api/urls/AbfdJ2",
		},
		{
			name:   "Delete",
			method: http.MethodDelete,
			path:   "/api/urls/AbfdJ2",
		},
		{
			name:   "Update",
			method: http.MethodPatch,
			path:   "/api/urls/AbfdJ2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handler.NewURLHandler(nil, logger)
			r := NewRouter(cfg, h, logger)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf(
					"expected status %d, got %d",
					http.StatusUnauthorized,
					w.Code,
				)
			}
		})
	}
}

func TestBasicAuth_Authorized(t *testing.T) {
	logger := fake.Logger

	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf(
			"env load failed: %v",
			err,
		)
	}

	cfg := config.Config{
		AppName: "url-shortener",
		HTTPServer: config.HTTPServer{
			User:     os.Getenv("BASIC_AUTH_USER"),
			Password: os.Getenv("BASIC_AUTH_PASSWORD"),
		},
	}

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{
			name:   "Save",
			method: http.MethodPost,
			path:   "/api/urls/",
			body:   `{"url":"https://example.com"}`,
			status: http.StatusCreated,
		},
		{
			name:   "Get",
			method: http.MethodGet,
			path:   "/api/urls/AbfdJ2",
			status: http.StatusOK,
		},
		{
			name:   "Delete",
			method: http.MethodDelete,
			path:   "/api/urls/AbfdJ2",
			status: http.StatusNoContent,
		},
		{
			name:   "Update",
			method: http.MethodPatch,
			path:   "/api/urls/AbfdJ2",
			body:   `{"url":"https://example.com/new"}`,
			status: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0

			result := service.URLResult{
				OriginalURL: "https://example.com",
				ShortURL:    "http://localhost:5050/AbfdJ2",
				Code:        "AbfdJ2",
			}

			mock := &mocks.MockService{
				SaveURLFunc: func(ctx context.Context, url string) (service.URLResult, error) {
					calls++
					return result, nil
				},

				GetURLByCodeFunc: func(ctx context.Context, code string) (service.URLResult, error) {
					calls++
					return result, nil
				},

				DeleteURLFunc: func(ctx context.Context, code string) error {
					calls++
					return nil
				},

				UpdateURLFunc: func(
					ctx context.Context,
					url *string,
					refresh bool,
					code string,
				) (service.URLResult, error) {
					calls++
					return result, nil
				},
			}

			h := handler.NewURLHandler(mock, logger)
			r := NewRouter(cfg, h, logger)

			req := httptest.NewRequest(
				tt.method,
				tt.path,
				strings.NewReader(tt.body),
			)

			req.SetBasicAuth(
				cfg.HTTPServer.User,
				cfg.HTTPServer.Password,
			)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.status {
				t.Errorf(
					"expected status %d, got %d",
					tt.status,
					w.Code,
				)
			}

			if calls != 1 {
				t.Errorf(
					"expected one service call, got %d",
					calls,
				)
			}
		})
	}
}
