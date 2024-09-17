package storage

import "context"

type Storage interface {
	Stop()

	WriteURL(ctx context.Context, origURL string) (shortURL string, err error)
	ReadURL(ctx context.Context, shortURL string) (origURL string, err error)
	Ping(ctx context.Context) error
	WriteURLs(ctx context.Context, origURLs []string) (shortURLs []string, err error)
}
