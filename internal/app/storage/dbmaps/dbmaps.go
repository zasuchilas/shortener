package dbmaps

import (
	"context"
	"errors"
	"fmt"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"github.com/zasuchilas/shortener/internal/app/storage/hashfuncs"
	"go.uber.org/zap"
	"sync"
	"time"
)

// DBMaps is a RAM storage on double maps
type DBMaps struct {
	urls   map[string]string
	hash   map[string]string
	lastID int64
	mutex  sync.RWMutex
}

func New() *DBMaps {
	db := &DBMaps{
		urls: make(map[string]string),
		hash: make(map[string]string),
	}
	return db
}

func (d *DBMaps) Stop() {}

func (d *DBMaps) WriteURL(ctx context.Context, origURL string) (shortURL string, conflict bool, err error) {
	logger.Log.Debug("checking if already exist")
	shortURL, ok := d.urls[origURL]
	if ok {
		return shortURL, true, nil
	}

	logger.Log.Debug("writing URL")
	urlRows, err := d.WriteURLs(ctx, []string{origURL})
	if err != nil {
		return "", false, err
	}
	if urlRows == nil || urlRows[origURL] == nil {
		return "", false, errors.New("something wrong with writing URL")
	}
	return urlRows[origURL].ShortURL, false, nil
}

func (d *DBMaps) ReadURL(_ context.Context, shortURL string) (origURL string, err error) {
	d.mutex.RLock()
	origURL, found := d.hash[shortURL]
	d.mutex.RUnlock()

	if !found {
		return "", errors.New("not found")
	}

	return origURL, nil
}

func (d *DBMaps) Ping(_ context.Context) error {
	return errors.New("not allowed")
}

func (d *DBMaps) WriteURLs(ctx context.Context, origURLs []string) (urlRows map[string]*models.URLRow, err error) {

	urlRows = make(map[string]*models.URLRow)

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	logger.Log.Debug("start ~tx in maps storage")
	d.mutex.Lock()
	defer d.mutex.Unlock()

loop:
	for _, origURL := range origURLs {
		select {
		case <-ctxTm.Done():
			err = fmt.Errorf("the operation was canceled")
			break loop
		default:
			logger.Log.Debug("default action")

			logger.Log.Debug("find is ready in maps storage", zap.String("origURL", origURL))
			shortURL, found := d.urls[origURL]
			if found {
				logger.Log.Debug("row already exist", zap.String("shortURL", shortURL))
				uuid, e := hashfuncs.DecodeZeroHash(shortURL)
				if e != nil {
					logger.Log.Error("DecodeZeroHash when row already exist", zap.Error(err))
					err = e
					break loop
				}
				urlRows[origURL] = &models.URLRow{
					UUID:     uuid,
					ShortURL: shortURL,
					OrigURL:  origURL,
				}
				continue
			}

			nextID := d.lastID + 1
			shortURL = hashfuncs.EncodeZeroHash(nextID)
			d.urls[origURL] = shortURL
			d.hash[shortURL] = origURL
			d.lastID = nextID
			logger.Log.Debug("inserted new row",
				zap.String("shortURL", shortURL), zap.String("origURL", origURL))

			urlRows[origURL] = &models.URLRow{
				UUID:     nextID,
				ShortURL: shortURL,
				OrigURL:  origURL,
			}
		}
	}
	if err != nil {
		return nil, err
	}

	return urlRows, nil
}
