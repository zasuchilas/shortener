package dbmaps

import (
	"context"
	"errors"
	"fmt"
	"github.com/zasuchilas/shortener/internal/app/logger"
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

//func (d *DBMaps) WriteURLs(ctx context.Context, origURL []string) (shortURLs []string, err error) {}

func (d *DBMaps) WriteURL(ctx context.Context, origURL string) (shortURL string, was bool, err error) {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	logger.Log.Debug("find is ready in maps storage", zap.String("origURL", origURL))
	v, found := d.urls[origURL]
	if found {
		return v, true, nil
	}

	logger.Log.Debug("start ~tx in maps storage", zap.String("origURL", origURL))
	d.mutex.Lock()
	defer d.mutex.Unlock()

	select {
	case <-ctxTm.Done():
		err = fmt.Errorf("the operation was canceled")
	default:
		logger.Log.Debug("default action")
		nextID := d.lastID + 1
		shortURL = hashfuncs.EncodeZeroHash(nextID)
		d.urls[origURL] = shortURL
		d.hash[shortURL] = origURL
		d.lastID = nextID
	}

	return shortURL, false, err
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
