package service

import (
	"context"

	"github.com/zasuchilas/shortener/internal/app/model"
)

// ShortenerService _
type ShortenerService interface {
	Ping(ctx context.Context) error
	ReadURL(ctx context.Context, shortURL string) (origURL string, err error)
	WriteURL(ctx context.Context, rawURL string, userID int64) (readyURL string, conflict bool, err error)
	ShortenBatch(ctx context.Context, in []model.ShortenBatchIn, userID int64) (out []model.ShortenBatchOut, err error)
	DeleteURLs(ctx context.Context, rawShortURLs []string, userID int64) error
	UserURLs(ctx context.Context, userID int64) (out []model.UserURL, err error)
	Stats(ctx context.Context) (out *model.Stats, err error)
}
