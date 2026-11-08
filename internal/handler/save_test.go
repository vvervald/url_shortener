package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"url-shortener/internal/lib/logger/fake"
	"url-shortener/internal/service"
)

func TestSave_BadRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "invalid JSON syntax",
			body: `{url:}`,
		},
		{
			name: "missing value",
			body: `{"url":}`,
		},
		{
			name: "unclosed string",
			body: `{"url":"http://test.com}`,
		},
		{
			name: "empty body",
			body: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/urls",
				strings.NewReader(tt.body),
			)

			w := httptest.NewRecorder()

			saveFunc := func(ctx context.Context, originalURL string) (service.URLResult, error) {
				t.Fatal("service must not be called for invalid JSON")
				return service.URLResult{}, nil
			}

			s := &mockService{
				saveURLFunc: saveFunc,
			}

			h := NewURLHandler(s, fake.Logger)

			h.Save(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf(
					"expected status code %d, got %d",
					http.StatusBadRequest,
					w.Code,
				)
			}

			var resp ErrorResponse

			err := json.Unmarshal(w.Body.Bytes(), &resp)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.Error != ErrInvalidJSON.Error() {
				t.Errorf(
					"expected error message %q, got: %q",
					ErrInvalidJSON.Error(),
					resp.Error,
				)
			}
		})
	}
}

func TestSave_Success(t *testing.T) {
	url := "https://test.com"
	body := fmt.Sprintf(`{"url":"%s"}`, url)
	code := "AbXfg3"
	shortURL := "http://localhost:5050/" + code
	calls := 0

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(body),
	)

	w := httptest.NewRecorder()

	saveFunc := func(
		ctx context.Context,
		originalURL string,
	) (service.URLResult, error) {
		calls += 1

		if originalURL != url {
			t.Errorf("expected url %q, got %q", url, originalURL)
		}

		return service.URLResult{
			OriginalURL: originalURL,
			ShortURL:    shortURL,
			Code:        code,
		}, nil
	}

	mockService := &mockService{
		saveURLFunc: saveFunc,
	}

	h := NewURLHandler(
		mockService,
		fake.Logger,
	)

	h.Save(w, req)

	if calls != 1 {
		t.Fatalf("service must be called once, number of calls %d", calls)
	}

	if w.Code != http.StatusCreated {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusCreated,
			w.Code,
		)
	}

	var resp URLResponse

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.OriginalURL != url {
		t.Errorf("expected original_url %q, got %q", url, resp.OriginalURL)
	}

	if resp.ShortURL != shortURL {
		t.Errorf("expected short_url %q, got %q", shortURL, resp.ShortURL)
	}

	if resp.Code != code {
		t.Errorf("expected code %q, got %q", code, resp.Code)
	}
}

func TestSave_InvalidURL(t *testing.T) {
	url := "htps://test.com"
	body := fmt.Sprintf(`{"url":"%s"}`, url)
	saveFunc := func(ctx context.Context, originalURL string) (service.URLResult, error) {
		if url != originalURL {
			t.Errorf(
				"expected url %q, got %q",
				url,
				originalURL,
			)
		}
		return service.URLResult{}, service.ErrInvalidURL
	}

	mockService := &mockService{
		saveURLFunc: saveFunc,
	}

	h := NewURLHandler(
		mockService,
		fake.Logger,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(body),
	)

	w := httptest.NewRecorder()

	h.Save(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusBadRequest,
			w.Code,
		)
	}

	var resp ErrorResponse

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unexpected errror: %v", err)
	}

	if resp.Error != ErrInvalidURL.Error() {
		t.Errorf("expected error message %q, got %q", ErrInvalidURL.Error(), resp.Error)
	}

}

func TestSave_InternalError(t *testing.T) {
	url := "https://test.com"
	body := fmt.Sprintf(`{"url":"%s"}`, url)

	saveFunc := func(ctx context.Context, originalURL string) (service.URLResult, error) {
		return service.URLResult{}, errors.New("unexpected error")
	}

	mockService := &mockService{
		saveURLFunc: saveFunc,
	}

	h := NewURLHandler(
		mockService,
		fake.Logger,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/urls",
		strings.NewReader(body),
	)

	w := httptest.NewRecorder()

	h.Save(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusInternalServerError,
			w.Code,
		)
	}

	var resp ErrorResponse

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf(
			"unexpected error %v",
			err,
		)
	}

	if resp.Error != ErrInternal.Error() {
		t.Errorf(
			"expected error message %q, got %q",
			ErrInternal.Error(),
			resp.Error,
		)
	}

}
