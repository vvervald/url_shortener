package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/lib/logger/fake"
	"url-shortener/internal/service"

	"github.com/go-chi/chi/v5"
)

func TestDelete_Success(t *testing.T) {
	code := "AdjfD0"

	deleteFunc := func(ctx context.Context, recievedCode string) error {
		if code != recievedCode {
			t.Errorf(
				"expected code %q, got %q",
				code,
				recievedCode,
			)
		}

		return nil
	}

	mockService := &mockService{
		deleteURLFunc: deleteFunc,
	}

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/urls/"+code,
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

	h := NewURLHandler(
		mockService,
		fake.Logger,
	)

	h.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusNoContent,
			w.Code,
		)
	}

	if w.Body.Len() != 0 {
		t.Errorf(
			"expected empty body, got %q",
			w.Body.String(),
		)
	}
}

func TestDelete_NotFound(t *testing.T) {
	code := "AdjfD0"

	deleteFunc := func(ctx context.Context, recievedCode string) error {
		if code != recievedCode {
			t.Errorf(
				"expected code %q, got %q",
				code,
				recievedCode,
			)
		}
		return service.ErrURLNotFound
	}

	mockService := &mockService{
		deleteURLFunc: deleteFunc,
	}

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/urls/"+code,
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

	h := NewURLHandler(
		mockService,
		fake.Logger,
	)

	h.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusNoContent,
			w.Code,
		)
	}

	if w.Body.Len() != 0 {
		t.Errorf(
			"expected empty body, got %q",
			w.Body.String(),
		)
	}
}
func TestDelete_InternalError(t *testing.T) {
	code := "AdjfD0"

	deleteFunc := func(ctx context.Context, recievedCode string) error {
		if code != recievedCode {
			t.Errorf(
				"expected code %q, got %q",
				code,
				recievedCode,
			)
		}
		return errors.New("unexpecteed error")
	}

	mockService := &mockService{
		deleteURLFunc: deleteFunc,
	}

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/urls/"+code,
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

	h := NewURLHandler(
		mockService,
		fake.Logger,
	)

	h.Delete(w, req)

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
