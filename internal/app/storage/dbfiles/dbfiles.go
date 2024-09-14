package dbfiles

import (
	"context"
	"errors"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/storage/hashfuncs"
	"go.uber.org/zap"
	"sync"
)

// DBFiles is a file storage implementation
type DBFiles struct {
	mutex sync.RWMutex
}

func New() *DBFiles {
	return &DBFiles{
		mutex: sync.RWMutex{},
	}
}

func (d *DBFiles) Stop() {}

func (d *DBFiles) WriteURL(_ context.Context, origURL string) (shortURL string, was bool, err error) {

	logger.Log.Debug("find in file storage", zap.String("origURL", origURL))
	//v, found, err := d.findOrig(origURL)
	v, found, err := d.lookup(nil, &origURL, nil)
	if err != nil {
		return "", false, err
	}
	if found {
		return v.ShortURL, true, nil
	}

	logger.Log.Debug("making short url", zap.String("origURL", origURL))
	shortURL, err = hashfuncs.MakeShortURL(d.isExist)
	if err != nil {
		return "", false, err
	}

	logger.Log.Debug("append to file storage", zap.String("origURL", origURL))
	err = d.writeRow(shortURL, origURL)
	if err != nil {
		return "", false, err
	}

	return shortURL, false, nil
}

func (d *DBFiles) ReadURL(_ context.Context, shortURL string) (origURL string, err error) {
	v, found, err := d.lookup(nil, nil, &shortURL)
	if err != nil {
		logger.Log.Error("error during lookup shortURL in file storage", zap.Error(err))
		return "", err // TODO: -> 500
	}
	if !found {
		return "", errors.New("not found")
	}

	return v.OrigURL, nil
}

func (d *DBFiles) Ping(_ context.Context) error {
	return errors.New("not allowed")
}
