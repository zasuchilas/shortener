package storage

import (
	"context"
	"github.com/zasuchilas/shortener/internal/app/models"
)

type Storage interface {
	Stop()

	WriteURL(ctx context.Context, origURL string) (shortURL string, conflict bool, err error)
	ReadURL(ctx context.Context, shortURL string) (origURL string, err error)
	Ping(ctx context.Context) error
	WriteURLs(ctx context.Context, origURLs []string) (urlRows map[string]*models.URLRow, err error)
}
