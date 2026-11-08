package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/lib/logger/fake"
	"url-shortener/internal/mocks"
	"url-shortener/internal/service"

	"github.com/go-chi/chi/v5"
)

func TestGet_Success(t *testing.T) {
	calls := 0
	code := "AsdJ89"
	originalURL := "http://example.com"
	shortURL := "localhost:5050/" + code

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/urls/"+code,
		nil,
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(
		"code", code,
	)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	w := httptest.NewRecorder()

	mockService := &mocks.MockService{
		GetURLByCodeFunc: func(ctx context.Context, recievedCode string) (service.URLResult, error) {
			calls++
			if recievedCode != code {
				t.Errorf(
					"expected code %q, got %q",
					code,
					recievedCode,
				)
			}

			return service.URLResult{
				OriginalURL: originalURL,
				ShortURL:    shortURL,
				Code:        code,
			}, nil
		},
	}

	h := NewURLHandler(
		mockService,
		fake.Logger,
	)

	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusOK,
			w.Code,
		)
	}

	var resp URLResponse

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf(
			"unexpected error %v",
			err,
		)
	}

	if resp.OriginalURL != originalURL {
		t.Errorf(
			"expected original url %q, got %q",
			originalURL,
			resp.OriginalURL,
		)
	}

	if resp.ShortURL != shortURL {
		t.Errorf(
			"expected short url %q, got %q",
			shortURL,
			resp.ShortURL,
		)
	}

	if resp.Code != code {
		t.Errorf(
			"expected code %q, got %q",
			code,
			resp.Code,
		)
	}
}

func TestGet_NotFound(t *testing.T) {
	calls := 0
	code := "AsdJ89"

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/urls/"+code,
		nil,
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(
		"code", code,
	)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	w := httptest.NewRecorder()

	mockService := &mocks.MockService{
		GetURLByCodeFunc: func(ctx context.Context, recievedCode string) (service.URLResult, error) {
			calls++
			if recievedCode != code {
				t.Errorf(
					"expected code %q, got %q",
					code,
					recievedCode,
				)
			}

			return service.URLResult{}, service.ErrURLNotFound
		},
	}

	h := NewURLHandler(
		mockService,
		fake.Logger,
	)

	h.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusNotFound,
			w.Code,
		)
	}

	var resp ErrorResponse

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf(
			"unexpected error %v",
			err,
		)
	}

	if resp.Error != ErrURLNotFound.Error() {
		t.Errorf(
			"expected error message %q, got %q",
			ErrURLNotFound.Error(),
			resp.Error,
		)
	}

}

func TestGet_InternalError(t *testing.T) {
	calls := 0
	code := "AsdJ89"

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/urls/"+code,
		nil,
	)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(
		"code", code,
	)

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			routeCtx,
		),
	)

	w := httptest.NewRecorder()

	mockService := &mocks.MockService{
		GetURLByCodeFunc: func(ctx context.Context, recievedCode string) (service.URLResult, error) {
			calls++
			if recievedCode != code {
				t.Errorf(
					"expected code %q, got %q",
					code,
					recievedCode,
				)
			}

			return service.URLResult{}, errors.New("unexpected error")
		},
	}

	h := NewURLHandler(
		mockService,
		fake.Logger,
	)

	h.Get(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusInternalServerError,
			w.Code,
		)
	}

	var resp ErrorResponse

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf(
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
