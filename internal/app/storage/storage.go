package storage

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"strings"
	"time"

	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
)

// Names implementation of the storage interface.
const (
	InstanceMemory     = "dbmaps"  // inmemory
	InstanceFile       = "dbfiles" // files
	InstancePostgresql = "dbpgsql" // postgresql
)

// Group deletion settings.
const (
	DeletingChanBuffer     = 1024
	DeletingMaxRowsRequest = 512
	DeletingFlushInterval  = 10 * time.Second
)

// Errors returned from the package.
var (
	ErrNotFound   = errors.New("not found")
	ErrGone       = errors.New("deleted")
	ErrBadRequest = errors.New("bad request")
)

// IStorage describes the interface to be implemented.
type IStorage interface {
	// Stop stops the component.
	Stop()

	// InstanceName returns current instance name.
	InstanceName() string

	// WriteURL writes URL in the storage.
	WriteURL(ctx context.Context, origURL string, userID int64) (shortURL string, conflict bool, err error)

	// ReadURL reads URL from the storage.
	ReadURL(ctx context.Context, shortURL string) (origURL string, err error)

	// Ping pings the storage.
	Ping(ctx context.Context) error

	// WriteURLs writes URLs in the storage.
	WriteURLs(ctx context.Context, origURLs []string, userID int64) (urlRows map[string]*models.URLRow, err error)

	// UserURLs returns user URLs from storage.
	UserURLs(ctx context.Context, userID int64) (urlRowList []*models.URLRow, err error)

	// CheckDeletedURLs checks deleting URLs.
	CheckDeletedURLs(ctx context.Context, userID int64, shortURLs []string) error

	// DeleteURLs deletes URLs from the storage.
	DeleteURLs(ctx context.Context, shortURLs ...string) error

	// Stats returns urls and users count.
	Stats(ctx context.Context) (int, error)
}

// checkUserURLs checks whether the user has the ability to delete the url data.
func checkUserURLs(userID int64, urlRows map[string]*models.URLRow) error {

	logger.Log.Debug("checking user urls", zap.Any("urls", urlRows))
	if len(urlRows) == 0 {
		return fmt.Errorf("%w nothing was found for the passed short urls", ErrBadRequest)
	}

	// check auth rights
	alreadyDeleted := make([]string, 0)
	forbiddenRows := make([]string, 0)
	for _, row := range urlRows {
		if row.UserID != userID {
			forbiddenRows = append(forbiddenRows, row.ShortURL)
		}
		if row.Deleted {
			alreadyDeleted = append(alreadyDeleted, row.ShortURL)
		}
	}
	if len(forbiddenRows) > 0 {
		return fmt.Errorf("%w you can't delete other people's short links (%s)", ErrBadRequest, strings.Join(forbiddenRows, ", "))
	}
	if len(alreadyDeleted) == len(urlRows) {
		return fmt.Errorf("%w all short urls have already been deleted", ErrBadRequest)
	}

	return nil
}
