package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
	"url-shortener/internal/model"
	"url-shortener/internal/repository"
)

var (
	ErrURLNotFound          = errors.New("URL not found")
	ErrEmptyURL             = errors.New("URL is required")
	ErrInvalidURL           = errors.New("invalid URL format")
	ErrCodeGenerationFailed = errors.New("code generation failed")
	ErrURLAlreadyExists     = errors.New("URL already exists")
	ErrNothingToUpdate      = errors.New("error nothing to update")
)

type URLResult struct {
	OriginalURL string
	Code        string
	ShortURL    string
}

type Service interface {
	SaveURL(ctx context.Context, originalURL string) (URLResult, error)
	DeleteURL(ctx context.Context, code string) error
	GetURLByCode(ctx context.Context, code string) (URLResult, error)
	UpdateURL(ctx context.Context, newURL *string, refreshCode bool, code string) (URLResult, error)
}

type URLService struct {
	baseURL             string
	codeLength          int
	aliasGenMaxAttempts int
	repo                repository.Repository
}

func NewURLService(
	baseURL string,
	codeLength int,
	aliasGenMaxAttempts int,
	repo repository.Repository,

) *URLService {
	return &URLService{
		baseURL:             baseURL,
		codeLength:          codeLength,
		aliasGenMaxAttempts: aliasGenMaxAttempts,
		repo:                repo,
	}
}

func (s *URLService) SaveURL(ctx context.Context, originalURL string) (URLResult, error) {
	if err := validateURL(originalURL); err != nil {
		return URLResult{}, err
	}

	for range s.aliasGenMaxAttempts {
		code, err := s.generateCode(s.codeLength)

		if err != nil {
			return URLResult{}, err
		}

		url, err := s.repo.Create(ctx, originalURL, code)

		if err != nil {
			switch {
			case errors.Is(err, repository.ErrCodeAlreadyExists):
				continue
			case errors.Is(err, repository.ErrURLAlreadyExists):
				url, err := s.repo.GetURLByOriginalURL(ctx, originalURL)

				if err == nil {
					return s.toResult(url), nil
				}

				return URLResult{}, err
			default:
				return URLResult{}, err
			}
		}

		return s.toResult(url), nil
	}

	return URLResult{}, ErrCodeGenerationFailed
}

func (s *URLService) DeleteURL(ctx context.Context, code string) error {
	if err := s.repo.DeleteURLByCode(ctx, code); err != nil {

		if errors.Is(err, repository.ErrNotFound) {
			return ErrURLNotFound
		}

		return err
	}

	return nil
}

func (s *URLService) GetURLByCode(ctx context.Context, code string) (URLResult, error) {
	url, err := s.repo.GetURLByCode(ctx, code)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return URLResult{}, ErrURLNotFound
		}
		return URLResult{}, err
	}

	return URLResult{
		OriginalURL: url.OriginalURL,
		Code:        url.Code,
		ShortURL:    s.buildShortURL(url.Code),
	}, nil
}

func (s *URLService) UpdateURL(ctx context.Context, newURL *string, refreshCode bool, code string) (URLResult, error) {
	if !refreshCode && newURL == nil {
		return URLResult{}, ErrNothingToUpdate
	}

	if newURL != nil {
		if err := validateURL(*newURL); err != nil {
			return URLResult{}, ErrInvalidURL
		}
	}

	if refreshCode {
		for range s.aliasGenMaxAttempts {
			newCode, err := s.generateCode(s.codeLength)

			if err != nil {
				return URLResult{}, err
			}

			url, err := s.repo.UpdateURL(ctx, newURL, &newCode, code)

			if err != nil {
				switch {
				case errors.Is(err, repository.ErrNotFound):
					return URLResult{}, ErrURLNotFound
				case errors.Is(err, repository.ErrURLAlreadyExists):
					return URLResult{}, ErrURLAlreadyExists
				case errors.Is(err, repository.ErrCodeAlreadyExists):
					continue
				default:
					return URLResult{}, err
				}
			}

			return s.toResult(url), nil
		}

		return URLResult{}, ErrCodeGenerationFailed
	}

	url, err := s.repo.UpdateURL(ctx, newURL, nil, code)

	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			return URLResult{}, ErrURLNotFound
		case errors.Is(err, repository.ErrURLAlreadyExists):
			return URLResult{}, ErrURLAlreadyExists
		default:
			return URLResult{}, err
		}
	}

	return s.toResult(url), nil
}

func (s *URLService) buildShortURL(code string) string {
	return s.baseURL + "/" + code
}

func (s *URLService) generateCode(length int) (string, error) {
	alphabet := "QWERTYUIOPASDFGHJKLZXCVBNMzxcvbnmasdfghjklqwertyuiop0987654321"
	bytes := make([]byte, 0, length)

	for range length {

		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))

		if err != nil {
			return "", err
		}

		bytes = append(bytes, alphabet[n.Int64()])

	}
	return string(bytes), nil
}

func validateURL(URL string) error {
	if URL == "" {
		return ErrEmptyURL
	}

	u, err := url.ParseRequestURI(URL)

	if err != nil {
		return ErrInvalidURL
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidURL
	}

	if u.Host == "" {
		return ErrInvalidURL
	}

	return nil
}

func (s *URLService) toResult(url model.URL) URLResult {
	return URLResult{
		OriginalURL: url.OriginalURL,
		Code:        url.Code,
		ShortURL:    s.buildShortURL(url.Code),
	}
}
