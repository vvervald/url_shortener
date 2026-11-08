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

	"github.com/go-chi/chi/v5"
)

func TestUpdate_Success(t *testing.T) {
	currentCode := "AbfdJ2"
	newCode := "Xyz123"
	originalURL := "https://example.com"
	newURL := "https://example.com/new"

	tests := []struct {
		name         string
		body         string
		expectedURL  string
		expectedCode string
		refreshCode  bool
		hasNewURL    bool
	}{
		{
			name:         "update original URL",
			body:         `{"url":"https://example.com/new"}`,
			expectedURL:  newURL,
			expectedCode: currentCode,
			refreshCode:  false,
			hasNewURL:    true,
		},
		{
			name:         "regenerate short URL",
			body:         `{"refresh_short_url":true}`,
			expectedURL:  originalURL,
			expectedCode: newCode,
			refreshCode:  true,
			hasNewURL:    false,
		},
		{
			name:         "update URL and regenerate code",
			body:         `{"url":"https://example.com/new","refresh_short_url":true}`,
			expectedURL:  newURL,
			expectedCode: newCode,
			refreshCode:  true,
			hasNewURL:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0

			updateFunc := func(
				ctx context.Context,
				receivedURL *string,
				receivedRefresh bool,
				receivedCode string,
			) (service.URLResult, error) {
				calls++

				if receivedCode != currentCode {
					t.Errorf(
						"expected code %q, got %q",
						currentCode,
						receivedCode,
					)
				}

				if receivedRefresh != tt.refreshCode {
					t.Errorf(
						"expected refreshCode %t, got %t",
						tt.refreshCode,
						receivedRefresh,
					)
				}

				if tt.hasNewURL {
					if receivedURL == nil {
						t.Error("expected non-nil newURL")
					} else if *receivedURL != newURL {
						t.Errorf(
							"expected newURL %q, got %q",
							newURL,
							*receivedURL,
						)
					}
				} else if receivedURL != nil {
					t.Errorf(
						"expected nil newURL, got %q",
						*receivedURL,
					)
				}

				return service.URLResult{
					OriginalURL: tt.expectedURL,
					Code:        tt.expectedCode,
					ShortURL:    "http://localhost:5050/" + tt.expectedCode,
				}, nil
			}

			mock := &mockService{
				updateURLFunc: updateFunc,
			}

			h := NewURLHandler(mock, fake.Logger)

			req := httptest.NewRequest(
				http.MethodPatch,
				"/api/urls/"+currentCode,
				strings.NewReader(tt.body),
			)

			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("code", currentCode)

			req = req.WithContext(
				context.WithValue(
					req.Context(),
					chi.RouteCtxKey,
					routeCtx,
				),
			)

			w := httptest.NewRecorder()

			h.Update(w, req)

			if calls != 1 {
				t.Errorf(
					"expected UpdateURL to be called once, got %d",
					calls,
				)
			}

			if w.Code != http.StatusOK {
				t.Errorf(
					"expected status %d, got %d",
					http.StatusOK,
					w.Code,
				)
			}

			var resp URLResponse

			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.OriginalURL != tt.expectedURL {
				t.Errorf(
					"expected original_url %q, got %q",
					tt.expectedURL,
					resp.OriginalURL,
				)
			}

			if resp.Code != tt.expectedCode {
				t.Errorf(
					"expected code %q, got %q",
					tt.expectedCode,
					resp.Code,
				)
			}

			expectedShortURL := "http://localhost:5050/" + tt.expectedCode

			if resp.ShortURL != expectedShortURL {
				t.Errorf(
					"expected short_url %q, got %q",
					expectedShortURL,
					resp.ShortURL,
				)
			}
		})
	}
}

func TestUpdate_BadRequest(t *testing.T) {
	code := "AbfdJ2"

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
			body: `{"url":"https://example.com}`,
		},
		{
			name: "empty body",
			body: "",
		},
		{
			name: "invalid field type",
			body: `{"refresh_short_url":"true"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPatch,
				"/api/urls/"+code,
				strings.NewReader(tt.body),
			)

			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("code", code)

			req = req.WithContext(
				context.WithValue(
					req.Context(),
					chi.RouteCtxKey,
					routeCtx,
				),
			)

			w := httptest.NewRecorder()

			updateFunc := func(
				ctx context.Context,
				newURL *string,
				refreshCode bool,
				receivedCode string,
			) (service.URLResult, error) {
				t.Fatal("UpdateURL must not be called for invalid JSON")
				return service.URLResult{}, nil
			}

			mock := &mockService{
				updateURLFunc: updateFunc,
			}

			h := NewURLHandler(mock, fake.Logger)

			h.Update(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf(
					"expected status code %d, got %d",
					http.StatusBadRequest,
					w.Code,
				)
			}

			var resp ErrorResponse

			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Error != ErrInvalidJSON.Error() {
				t.Errorf(
					"expected error message %q, got %q",
					ErrInvalidJSON.Error(),
					resp.Error,
				)
			}
		})
	}
}

func TestUpdate_InvalidURL(t *testing.T) {
	code := "AbfdJ2"
	invalidURL := "htps://example.com"

	body := fmt.Sprintf(`{"url":"%s"}`, invalidURL)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/api/urls/"+code,
		strings.NewReader(body),
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("code", code)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	w := httptest.NewRecorder()

	calls := 0

	updateFunc := func(
		ctx context.Context,
		receivedURL *string,
		refreshCode bool,
		receivedCode string,
	) (service.URLResult, error) {
		calls++

		if receivedCode != code {
			t.Errorf(
				"expected code %q, got %q",
				code,
				receivedCode,
			)
		}

		if receivedURL == nil {
			t.Fatal("expected non-nil URL")
		}

		if *receivedURL != invalidURL {
			t.Errorf(
				"expected URL %q, got %q",
				invalidURL,
				*receivedURL,
			)
		}

		if refreshCode {
			t.Error("expected refreshCode to be false")
		}

		return service.URLResult{}, service.ErrInvalidURL
	}

	mock := &mockService{
		updateURLFunc: updateFunc,
	}

	h := NewURLHandler(mock, fake.Logger)

	h.Update(w, req)

	if calls != 1 {
		t.Errorf(
			"expected UpdateURL to be called once, got %d",
			calls,
		)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusBadRequest,
			w.Code,
		)
	}

	var resp ErrorResponse

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != ErrInvalidURL.Error() {
		t.Errorf(
			"expected error message %q, got %q",
			ErrInvalidURL.Error(),
			resp.Error,
		)
	}
}

func TestUpdate_NothingToUpdate(t *testing.T) {
	code := "AbfdJ2"
	body := `{}`

	req := httptest.NewRequest(
		http.MethodPatch,
		"/api/urls/"+code,
		strings.NewReader(body),
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("code", code)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	w := httptest.NewRecorder()

	calls := 0

	updateFunc := func(
		ctx context.Context,
		receivedURL *string,
		refreshCode bool,
		receivedCode string,
	) (service.URLResult, error) {
		calls++

		if receivedCode != code {
			t.Errorf(
				"expected code %q, got %q",
				code,
				receivedCode,
			)
		}

		if receivedURL != nil {
			t.Errorf(
				"expected nil URL, got %q",
				*receivedURL,
			)
		}

		if refreshCode {
			t.Error("expected refreshCode to be false")
		}

		return service.URLResult{}, service.ErrNothingToUpdate
	}

	mock := &mockService{
		updateURLFunc: updateFunc,
	}

	h := NewURLHandler(mock, fake.Logger)

	h.Update(w, req)

	if calls != 1 {
		t.Errorf(
			"expected UpdateURL to be called once, got %d",
			calls,
		)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusBadRequest,
			w.Code,
		)
	}

	var resp ErrorResponse

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != service.ErrNothingToUpdate.Error() {
		t.Errorf(
			"expected error message %q, got %q",
			service.ErrNothingToUpdate.Error(),
			resp.Error,
		)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	code := "AbfdJ2"
	newURL := "https://example.com/new"

	body := fmt.Sprintf(`{"url":"%s"}`, newURL)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/api/urls/"+code,
		strings.NewReader(body),
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("code", code)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	w := httptest.NewRecorder()

	calls := 0

	updateFunc := func(
		ctx context.Context,
		receivedURL *string,
		refreshCode bool,
		receivedCode string,
	) (service.URLResult, error) {
		calls++

		if receivedCode != code {
			t.Errorf(
				"expected code %q, got %q",
				code,
				receivedCode,
			)
		}

		if receivedURL == nil {
			t.Error("expected non-nil URL")
		} else if *receivedURL != newURL {
			t.Errorf(
				"expected URL %q, got %q",
				newURL,
				*receivedURL,
			)
		}

		if refreshCode {
			t.Error("expected refreshCode to be false")
		}

		return service.URLResult{}, service.ErrURLNotFound
	}

	mock := &mockService{
		updateURLFunc: updateFunc,
	}

	h := NewURLHandler(mock, fake.Logger)

	h.Update(w, req)

	if calls != 1 {
		t.Errorf(
			"expected UpdateURL to be called once, got %d",
			calls,
		)
	}

	if w.Code != http.StatusNotFound {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusNotFound,
			w.Code,
		)
	}

	var resp ErrorResponse

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != ErrURLNotFound.Error() {
		t.Errorf(
			"expected error message %q, got %q",
			ErrURLNotFound.Error(),
			resp.Error,
		)
	}
}

func TestUpdate_InternalError(t *testing.T) {
	code := "AbfdJ2"
	newURL := "https://example.com/new"
	body := fmt.Sprintf(`{"url":"%s"}`, newURL)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/api/urls/"+code,
		strings.NewReader(body),
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("code", code)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	w := httptest.NewRecorder()

	calls := 0
	internalErr := errors.New("database connection failed")

	updateFunc := func(
		ctx context.Context,
		receivedURL *string,
		refreshCode bool,
		receivedCode string,
	) (service.URLResult, error) {
		calls++

		if receivedCode != code {
			t.Errorf("expected code %q, got %q", code, receivedCode)
		}

		if receivedURL == nil {
			t.Error("expected non-nil URL")
		} else if *receivedURL != newURL {
			t.Errorf("expected URL %q, got %q", newURL, *receivedURL)
		}

		if refreshCode {
			t.Error("expected refreshCode to be false")
		}

		return service.URLResult{}, internalErr
	}

	mock := &mockService{
		updateURLFunc: updateFunc,
	}

	h := NewURLHandler(mock, fake.Logger)

	h.Update(w, req)

	if calls != 1 {
		t.Errorf("expected one UpdateURL call, got %d", calls)
	}

	if w.Code != http.StatusInternalServerError {
		t.Errorf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			w.Code,
		)
	}

	var resp ErrorResponse

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Error != ErrInternal.Error() {
		t.Errorf(
			"expected error %q, got %q",
			ErrInternal.Error(),
			resp.Error,
		)
	}
}
