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

func TestRedirect_Success(t *testing.T) {
	code := "AbfdJ2"
	redirectURL := "http://test.com"
	calls := 0

	req := httptest.NewRequest(
		http.MethodGet,
		"/"+code,
		nil,
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

	getURLfunc := func(ctx context.Context, recievedCode string) (service.URLResult, error) {
		calls++
		if recievedCode != code {
			t.Errorf(
				"expected code %q, got %q",
				code,
				recievedCode,
			)
		}

		return service.URLResult{
			OriginalURL: redirectURL,
			ShortURL:    "localhost:5050/" + code,
			Code:        code,
		}, nil
	}

	mockService := &mocks.MockService{
		GetURLByCodeFunc: getURLfunc,
	}

	h := NewURLHandler(
		mockService,
		fake.Logger,
	)

	h.Redirect(w, req)

	if calls != 1 {
		t.Fatalf("service must be called once, number of calls %d", calls)
	}

	if w.Code != http.StatusFound {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusFound,
			w.Code,
		)
	}

	location := w.Header().Get("Location")
	if location != redirectURL {
		t.Errorf(
			"expected redirect_url %q, got %q",
			redirectURL,
			location,
		)
	}
}

func TestRedirect_NotFound(t *testing.T) {
	code := "AbfdJ2"
	calls := 0

	req := httptest.NewRequest(
		http.MethodGet,
		"/"+code,
		nil,
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

	getURLfunc := func(ctx context.Context, recievedCode string) (service.URLResult, error) {
		calls++
		if recievedCode != code {
			t.Errorf(
				"expected code %q, got %q",
				code,
				recievedCode,
			)
		}

		return service.URLResult{}, service.ErrURLNotFound
	}

	mockService := &mocks.MockService{
		GetURLByCodeFunc: getURLfunc,
	}

	h := NewURLHandler(
		mockService,
		fake.Logger,
	)

	h.Redirect(w, req)

	if calls != 1 {
		t.Fatalf("service must be called once, number of calls %d", calls)
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
		t.Fatalf("unexpected error %v", err)
	}

	if resp.Error != ErrURLNotFound.Error() {
		t.Errorf(
			"expected error message %q, got %q",
			ErrURLNotFound.Error(),
			resp.Error,
		)
	}
}

func TestRedirect_InternalError(t *testing.T) {
	code := "AbfdJ2"
	calls := 0

	req := httptest.NewRequest(
		http.MethodGet,
		"/"+code,
		nil,
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

	getURLfunc := func(ctx context.Context, recievedCode string) (service.URLResult, error) {
		calls++
		if recievedCode != code {
			t.Errorf(
				"expected code %q, got %q",
				code,
				recievedCode,
			)
		}

		return service.URLResult{}, errors.New("unexpected error")
	}

	mockService := &mocks.MockService{
		GetURLByCodeFunc: getURLfunc,
	}

	h := NewURLHandler(
		mockService,
		fake.Logger,
	)

	h.Redirect(w, req)

	if calls != 1 {
		t.Errorf(
			"expected service to be called once, got %d",
			calls,
		)
	}

	if w.Code != http.StatusInternalServerError {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusInternalServerError,
			w.Code,
		)
	}

	var resp ErrorResponse

	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if resp.Error != ErrInternal.Error() {
		t.Errorf(
			"expected error message %q, got %q",
			ErrInternal.Error(),
			resp.Error,
		)
	}
}
