package dbmaps

import (
	"context"
	"errors"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/storage/hashfuncs"
	"go.uber.org/zap"
	"sync"
)

// DBMaps is a RAM storage on double maps
type DBMaps struct {
	urls  map[string]string
	hash  map[string]string
	mutex sync.RWMutex
}

func New() *DBMaps {
	db := &DBMaps{
		urls: make(map[string]string),
		hash: make(map[string]string),
	}
	return db
}

func (d *DBMaps) Stop() {}

func (d *DBMaps) WriteURL(_ context.Context, origURL string) (shortURL string, was bool, err error) {

	logger.Log.Debug("find in maps storage", zap.String("origURL", origURL))
	d.mutex.RLock()
	v, found := d.urls[origURL]
	d.mutex.RUnlock()
	if found {
		return v, true, nil
	}

	shortURL, err = hashfuncs.MakeShortURL(d.isExist)
	if err != nil {
		return "", false, err
	}

	logger.Log.Debug("append to maps storage")
	d.mutex.Lock()
	d.urls[origURL] = shortURL
	d.hash[shortURL] = origURL
	d.mutex.Unlock()

	return shortURL, false, nil
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
