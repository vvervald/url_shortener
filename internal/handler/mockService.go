package handler

import (
	"context"
	"url-shortener/internal/service"
)

// type Service interface {
// 	SaveURL(ctx context.Context, originalURL string) (URLResult, error)
// 	DeleteURL(ctx context.Context, code string) error
// 	GetURLByCode(ctx context.Context, code string) (URLResult, error)
// 	UpdateURL(ctx context.Context, newURL *string, refreshCode bool, code string) (URLResult, error)
// }

type mockService struct {
	service.Service

	saveURLFunc   func(ctx context.Context, originalURL string) (service.URLResult, error)
	deleteURLFunc func(ctx context.Context, code string) error
	getURLByCode  func(ctx context.Context, code string) (service.URLResult, error)
	updateURLFunc func(
		ctx context.Context,
		newURL *string,
		refreshCode bool,
		code string,
	) (service.URLResult, error)
}

func (s *mockService) SaveURL(ctx context.Context, originalURL string) (service.URLResult, error) {
	return s.saveURLFunc(ctx, originalURL)
}

func (s *mockService) DeleteURL(ctx context.Context, code string) error {
	return s.deleteURLFunc(ctx, code)
}

func (s *mockService) GetURLByCode(ctx context.Context, code string) (service.URLResult, error) {
	return s.getURLByCode(ctx, code)
}

func (m *mockService) UpdateURL(
	ctx context.Context,
	newURL *string,
	refreshCode bool,
	code string,
) (service.URLResult, error) {
	return m.updateURLFunc(ctx, newURL, refreshCode, code)
}
