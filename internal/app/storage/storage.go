package storage

import (
	"context"
	"errors"
	"github.com/zasuchilas/shortener/internal/app/models"
)

const (
	InstanceMemory     = "dbmaps"
	InstanceFile       = "dbfiles"
	InstancePostgresql = "dbpgsql"
)

var ErrNotFound = errors.New("not found")

type Storage interface {
	Stop()
	InstanceName() string

	WriteURL(ctx context.Context, origURL string, userID int64) (shortURL string, conflict bool, err error)
	ReadURL(ctx context.Context, shortURL string) (origURL string, err error)
	Ping(ctx context.Context) error
	WriteURLs(ctx context.Context, origURLs []string, userID int64) (urlRows map[string]*models.URLRow, err error)
	UserURLs(ctx context.Context, userID int64) (urlRowList []*models.URLRow, err error)
}
