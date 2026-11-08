package model

import "time"

type URL struct {
	ID          int64
	OriginalURL string
	Code        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
