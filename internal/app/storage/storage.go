package storage

import "context"

type Storage interface {
	Stop()

	WriteURL(ctx context.Context, origURL string) (shortURL string, was bool, err error)
	ReadURL(ctx context.Context, shortURL string) (origURL string, err error)
	Ping(ctx context.Context) error
}
