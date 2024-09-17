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

func (d *DBFiles) WriteURL(ctx context.Context, origURL string) (shortURL string, err error) {
	shortURLs, err := d.WriteURLs(ctx, []string{origURL})
	if err != nil {
		return "", err
	}
	if len(shortURLs) != 1 {
		return "", errors.New("something wrong with writing URL")
	}
	return shortURLs[0], nil
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

func (d *DBFiles) WriteURLs(ctx context.Context, origURLs []string) (shortURLs []string, err error) {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	logger.Log.Debug("start ~tx in file storage")
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

			logger.Log.Debug("find is ready in file storage", zap.String("origURL", origURL))
			shortURL, found := d.urls[origURL]
			if found {
				logger.Log.Debug("row already exist", zap.String("shortURL", shortURL))
				shortURLs = append(shortURLs, shortURL)
				continue
			}

			nextID := d.lastID + 1
			shortURL = hashfuncs.EncodeZeroHash(nextID)
			err = d.writeRow(nextID, shortURL, origURL)
			if err != nil {
				break loop
			}
			d.urls[origURL] = shortURL
			d.hash[shortURL] = origURL
			d.lastID = nextID
			shortURLs = append(shortURLs, shortURL)
			logger.Log.Debug("inserted new row",
				zap.String("shortURL", shortURL), zap.String("origURL", origURL))
		}
	}
	if err != nil {
		return nil, err
	}

	return shortURLs, nil
}
