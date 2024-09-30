package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"go.uber.org/zap"
	"strings"
	"time"
)

const (
	InstanceMemory        = "dbmaps"
	InstanceFile          = "dbfiles"
	InstancePostgresql    = "dbpgsql"
	DeletingChanBuffer    = 1024
	DeletingFlushInterval = 10 * time.Second
)

var (
	ErrNotFound   = errors.New("not found")
	ErrGone       = errors.New("deleted")
	ErrBadRequest = errors.New("bad request")
)

type Storage interface {
	Stop()
	InstanceName() string

	WriteURL(ctx context.Context, origURL string, userID int64) (shortURL string, conflict bool, err error)
	ReadURL(ctx context.Context, shortURL string) (origURL string, err error)
	Ping(ctx context.Context) error
	WriteURLs(ctx context.Context, origURLs []string, userID int64) (urlRows map[string]*models.URLRow, err error)
	UserURLs(ctx context.Context, userID int64) (urlRowList []*models.URLRow, err error)
	CheckDeletedURLs(ctx context.Context, userID int64, shortURLs []string) error
	DeleteURLs(ctx context.Context, shortURLs ...string) error
}

func checkUserURLs(userID int64, urlRows map[string]*models.URLRow) error {

	logger.Log.Debug("checking user urls", zap.Any("urls", urlRows))
	if urlRows == nil || len(urlRows) == 0 {
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
