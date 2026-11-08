package mocks

import (
	"context"
	"url-shortener/internal/service"
)

type MockService struct {
	SaveURLFunc      func(ctx context.Context, originalURL string) (service.URLResult, error)
	DeleteURLFunc    func(ctx context.Context, code string) error
	GetURLByCodeFunc func(ctx context.Context, code string) (service.URLResult, error)
	UpdateURLFunc    func(
		ctx context.Context,
		newURL *string,
		refreshCode bool,
		code string,
	) (service.URLResult, error)
}

func (s *MockService) SaveURL(ctx context.Context, originalURL string) (service.URLResult, error) {
	return s.SaveURLFunc(ctx, originalURL)
}

func (s *MockService) DeleteURL(ctx context.Context, code string) error {
	return s.DeleteURLFunc(ctx, code)
}

func (s *MockService) GetURLByCode(ctx context.Context, code string) (service.URLResult, error) {
	return s.GetURLByCodeFunc(ctx, code)
}

func (m *MockService) UpdateURL(
	ctx context.Context,
	newURL *string,
	refreshCode bool,
	code string,
) (service.URLResult, error) {
	return m.UpdateURLFunc(ctx, newURL, refreshCode, code)
}
