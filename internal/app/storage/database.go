package storage

import (
	"errors"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"go.uber.org/zap"
	"sync"
)

type Database struct {
	Urls  map[string]string
	Hash  map[string]string
	mutex sync.RWMutex
}

func New() Storage {
	db := &Database{
		Urls: make(map[string]string),
		Hash: make(map[string]string),
	}
	restoreFromFile(db, config.FileStoragePath)
	return db
}

func restoreFromFile(db *Database, filename string) {

	logger.Log.Debug("checking filename")
	if filename == "" {
		return
	}

	logger.Log.Debug("making consumer for storage")
	consumer, err := NewConsumer(config.FileStoragePath)
	if err != nil {
		logger.Log.Fatal("making consumer for storage", zap.Error(err))
	}
	defer consumer.Close()

	logger.Log.Debug("populating cache storage from file storage")
	for {
		row, err := consumer.ReadURLRow()
		if err != nil {
			logger.Log.Debug("reading url from file", zap.Error(err))
			break
		}
		db.Urls[row.OriginalURL] = row.ShortURL
		db.Hash[row.ShortURL] = row.OriginalURL
	}

}

func (d *Database) WriteURL(rawURL string) (shortURL string, err error) {

	u, err := d.cleanURL(rawURL)
	if err != nil {
		return "", err
	}

	// find in storage
	v, found := d.Urls[u]
	if found {
		return v, nil
	}

	shortURL, err = d.makeShortURL()
	if err != nil {
		return "", err
	}

	// write to storage
	d.mutex.Lock()
	logger.Log.Debug("append to map (cache) storage")
	d.Urls[u] = shortURL
	d.Hash[shortURL] = u

	logger.Log.Debug("append to file storage")
	logger.Log.Debug("making producer for storage")
	producer, err := NewProducer(config.FileStoragePath)
	if err != nil {
		logger.Log.Fatal("making producer for storage", zap.Error(err))
	}
	defer producer.Close()

	err = producer.WriteURLRow(shortURL, u)
	if err != nil {
		logger.Log.Debug("error append row into file storage", zap.Error(err))
	}
	d.mutex.Unlock()

	return shortURL, nil
}

func (d *Database) ReadURL(shortURL string) (origURL string, err error) {
	d.mutex.RLock()
	origURL, found := d.Hash[shortURL]
	d.mutex.RUnlock()

	if !found {
		return "", errors.New("not found")
	}

	return origURL, nil
}
