package repository

import (
	"context"
	"errors"
	"url-shortener/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound          error = errors.New("url not found")
	ErrCodeAlreadyExists error = errors.New("such code already exists")
	ErrURLAlreadyExists  error = errors.New("such url already exists")
)

type Repository interface {
	Create(ctx context.Context, originalURL string, code string) (model.URL, error)
	DeleteURLByCode(ctx context.Context, code string) error
	GetURLByCode(ctx context.Context, code string) (model.URL, error)
	GetURLByOriginalURL(ctx context.Context, originalURL string) (model.URL, error)
	UpdateURL(ctx context.Context, newURL *string, newCode *string, currentCode string) (model.URL, error)
}

type URLRepository struct {
	db *pgxpool.Pool
}

func NewURLRepository(db *pgxpool.Pool) *URLRepository {
	return &URLRepository{
		db: db,
	}
}

func (r *URLRepository) Create(ctx context.Context, originalURL, code string) (model.URL, error) {
	row := r.db.QueryRow(
		ctx,
		`
		INSERT INTO urls(original_url, code)
		VALUES ($1, $2)
		RETURNING id, original_url, code, created_at; 
		`,
		originalURL,
		code,
	)

	var url model.URL
	var pgErr *pgconn.PgError

	if err := row.Scan(
		&url.ID,
		&url.OriginalURL,
		&url.Code,
		&url.CreatedAt,
	); err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "urls_original_url_unique":
				return url, ErrURLAlreadyExists
			case "urls_code_unique":
				return url, ErrCodeAlreadyExists
			}
		}
		return url, err
	}

	return url, nil
}

func (r *URLRepository) GetURLByCode(ctx context.Context, code string) (model.URL, error) {
	row := r.db.QueryRow(ctx,
		`
		SELECT id, original_url, code, created_at
		FROM urls 
		WHERE code=$1;  
		`,
		code,
	)

	var url model.URL

	if err := row.Scan(
		&url.ID,
		&url.OriginalURL,
		&url.Code,
		&url.CreatedAt,
	); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return url, ErrNotFound
		}
		return url, err
	}

	return url, nil
}

func (r *URLRepository) DeleteURLByCode(ctx context.Context, code string) error {
	_, err := r.db.Exec(ctx,
		`
		DELETE FROM urls
		WHERE code=$1;
	`,
		code,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *URLRepository) GetURLByOriginalURL(ctx context.Context, originalURL string) (model.URL, error) {
	row := r.db.QueryRow(ctx,
		`
		SELECT id, original_url, code, created_at 
		FROM urls 
		WHERE original_url=$1;
	`,
		originalURL,
	)

	var url model.URL

	if err := row.Scan(
		&url.ID,
		&url.OriginalURL,
		&url.Code,
		&url.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return url, ErrNotFound
		}

		return url, err
	}

	return url, nil
}

func (r *URLRepository) UpdateURL(ctx context.Context, newURL *string, newCode *string, currentCode string) (model.URL, error) {
	var url model.URL
	sql := `
		UPDATE urls 
		SET 	
			original_url=COALESCE($1, original_url),
			code=COALESCE($2, code), 
			updated_at=NOW()
		WHERE code=$3
		RETURNING id, original_url, code, created_at, updated_at 
	`
	row := r.db.QueryRow(
		ctx,
		sql,
		newURL,
		newCode,
		currentCode,
	)

	if err := row.Scan(
		&url.ID,
		&url.OriginalURL,
		&url.Code,
		&url.CreatedAt,
		&url.UpdatedAt,
	); err != nil {

		var pgxErr *pgconn.PgError

		if errors.Is(err, pgx.ErrNoRows) {
			return model.URL{}, ErrNotFound
		}

		if errors.As(err, &pgxErr) && pgxErr.Code == "23505" {
			switch pgxErr.ConstraintName {
			case "urls_original_url_key":
				return model.URL{}, ErrURLAlreadyExists
			case "urls_code_key":
				return model.URL{}, ErrCodeAlreadyExists
			}
		}

		return model.URL{}, err
	}

	return url, nil
}
