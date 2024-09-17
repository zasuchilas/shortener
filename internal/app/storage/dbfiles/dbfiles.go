package dbfiles

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

// DBFiles is a file storage implementation
type DBFiles struct {
	urls   map[string]string
	hash   map[string]string
	lastID int64
	mutex  sync.RWMutex
}

func New() *DBFiles {
	db := &DBFiles{
		urls:  make(map[string]string),
		hash:  make(map[string]string),
		mutex: sync.RWMutex{},
	}

	lastID, err := db.loadFromFile()
	if err != nil {
		logger.Log.Fatal("loading data from file", zap.Error(err))
	}
	db.lastID = lastID

	return db
}

func (d *DBFiles) Stop() {}

func (d *DBFiles) WriteURL(ctx context.Context, origURL string) (shortURL string, was bool, err error) {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	logger.Log.Debug("find is ready in file storage", zap.String("origURL", origURL))
	v, found := d.urls[origURL]
	if found {
		return v, true, nil
	}

	logger.Log.Debug("start ~tx in file storage", zap.String("origURL", origURL))
	d.mutex.Lock()
	defer d.mutex.Unlock()

	select {
	case <-ctxTm.Done():
		return "", false, fmt.Errorf("the operation was canceled")
	default:
		logger.Log.Debug("default action")
		nextID := d.lastID + 1
		shortURL = hashfuncs.EncodeZeroHash(nextID)
		err = d.writeRow(nextID, shortURL, origURL)
		if err != nil {
			return "", false, err
		}
		d.urls[origURL] = shortURL
		d.hash[shortURL] = origURL
		d.lastID = nextID
	}

	return shortURL, false, nil
}

func (d *DBFiles) ReadURL(_ context.Context, shortURL string) (origURL string, err error) {
	d.mutex.RLock()
	origURL, found := d.hash[shortURL]
	d.mutex.RUnlock()

	if !found {
		return "", errors.New("not found")
	}

	return origURL, nil
}

func (d *DBFiles) Ping(_ context.Context) error {
	return errors.New("not allowed")
}
