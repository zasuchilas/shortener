package dbfiles

import (
	"context"
	"errors"
	"fmt"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"github.com/zasuchilas/shortener/internal/app/storage/hashfuncs"
	"github.com/zasuchilas/shortener/internal/utils/filefuncs"
	"go.uber.org/zap"
	"io"
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

func (d *DBFiles) WriteURL(ctx context.Context, origURL string) (shortURL string, conflict bool, err error) {
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

func (d *DBFiles) WriteURLs(ctx context.Context, origURLs []string) (urlRows map[string]*models.URLRow, err error) {

	urlRows = make(map[string]*models.URLRow)

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
			err = d.writeRow(nextID, shortURL, origURL)
			if err != nil {
				logger.Log.Error("writing new row to file", zap.Error(err))
				break loop
			}
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

// TODO: as an option: use cache lib with reading from file

func (d *DBFiles) loadFromFile() (lastID int64, err error) {
	r, err := filefuncs.NewFileReader(config.FileStoragePath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	var lastHash string
	for {
		row, e := r.ReadURLRow()
		if e == io.EOF {
			break
		}
		if e != nil {
			logger.Log.Debug("reading urls from file", zap.Error(err))
			err = e
			break
		}

		d.urls[row.OrigURL] = row.ShortURL
		d.hash[row.ShortURL] = row.OrigURL
		lastHash = row.ShortURL
	}

	if lastHash == "" {
		return 0, nil
	}

	lastID, err = hashfuncs.DecodeZeroHash(lastHash)
	if err != nil {
		return 0, err
	}

	return lastID, nil
}

func (d *DBFiles) writeRow(uuid int64, shortURL, origURL string) error {
	w, err := filefuncs.NewFileWriter(config.FileStoragePath)
	if err != nil {
		return err
	}
	defer w.Close()

	return w.WriteURLRow(uuid, shortURL, origURL)
}
