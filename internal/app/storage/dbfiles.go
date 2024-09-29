package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"github.com/zasuchilas/shortener/internal/app/utils/filefuncs"
	"github.com/zasuchilas/shortener/internal/app/utils/hashfuncs"
	"go.uber.org/zap"
	"io"
	"sync"
	"time"
)

// DBFiles is a file storage implementation
type DBFiles struct {
	urls   map[string]*models.URLRow
	hash   map[string]*models.URLRow
	owners map[int64][]*models.URLRow
	lastID int64
	mutex  sync.RWMutex
}

func NewDBFile() *DBFiles {
	db := &DBFiles{
		urls:   make(map[string]*models.URLRow),
		hash:   make(map[string]*models.URLRow),
		owners: make(map[int64][]*models.URLRow),
		mutex:  sync.RWMutex{},
	}

	lastID, err := db.loadFromFile()
	if err != nil {
		logger.Log.Fatal("loading data from file", zap.Error(err))
	}
	db.lastID = lastID

	return db
}

func (d *DBFiles) Stop() {}

func (d *DBFiles) InstanceName() string {
	return InstanceFile
}

func (d *DBFiles) WriteURL(ctx context.Context, origURL string, userID int64) (shortURL string, conflict bool, err error) {
	// checking if already exist
	found, ok := d.urls[origURL]
	if ok {
		return found.ShortURL, true, nil
	}

	// writing URL
	urlRows, err := d.WriteURLs(ctx, []string{origURL}, userID)
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
	found, ok := d.hash[shortURL]
	d.mutex.RUnlock()

	if !ok {
		return "", fmt.Errorf("%w", ErrNotFound)
	}

	if found.Deleted {
		return "", fmt.Errorf("%w", ErrGone)
	}

	return found.OrigURL, nil
}

func (d *DBFiles) Ping(_ context.Context) error {
	return errors.New("not allowed")
}

func (d *DBFiles) WriteURLs(ctx context.Context, origURLs []string, userID int64) (urlRows map[string]*models.URLRow, err error) {

	urlRows = make(map[string]*models.URLRow)

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// start ~tx in file storage
	d.mutex.Lock()
	defer d.mutex.Unlock()

	w, err := filefuncs.NewFileWriter(config.FileStoragePath)
	if err != nil {
		return nil, err
	}
	defer w.Close()

	// if new user (he doesn't have his data)
	_, ex := d.owners[userID]
	if !ex {
		d.owners[userID] = make([]*models.URLRow, 0)
	}

loop:
	for _, origURL := range origURLs {
		select {
		case <-ctxTm.Done():
			err = fmt.Errorf("the operation was canceled")
			break loop
		default:
			logger.Log.Debug("find is ready in file storage", zap.String("origURL", origURL))
			found, ok := d.urls[origURL]
			if ok {
				logger.Log.Debug("row already exist", zap.String("shortURL", found.ShortURL))
				urlRows[origURL] = found
				continue
			}

			nextID := d.lastID + 1
			shortURL := hashfuncs.EncodeZeroHash(nextID)
			nextURLRow := &models.URLRow{
				ID:       nextID,
				ShortURL: shortURL,
				OrigURL:  origURL,
				UserID:   userID,
				Deleted:  false,
			}

			// writing new row to file storage
			err = w.WriteURLRow(nextURLRow)
			if err != nil {
				logger.Log.Error("writing new row to file", zap.Error(err))
				break loop
			}

			d.urls[origURL] = nextURLRow
			d.hash[shortURL] = nextURLRow
			d.owners[userID] = append(d.owners[userID], nextURLRow)
			d.lastID = nextID

			logger.Log.Debug("inserted new row",
				zap.String("shortURL", shortURL), zap.String("origURL", origURL))
			urlRows[origURL] = nextURLRow
		}
	}
	if err != nil {
		return nil, err
	}

	return urlRows, nil
}

func (d *DBFiles) UserURLs(_ context.Context, userID int64) (urlRowList []*models.URLRow, err error) {

	d.mutex.RLock()
	found, ok := d.owners[userID]
	d.mutex.RUnlock()

	if !ok || len(found) == 0 {
		return nil, fmt.Errorf("%w", ErrNotFound)
	}

	return found, nil
}

func (d *DBFiles) CheckDeletedURLs(_ context.Context, userID int64, shortURLs []string) error {
	d.mutex.RLock()
	found, ok := d.owners[userID]
	d.mutex.RUnlock()

	if !ok || len(found) == 0 {
		return checkUserURLs(userID, nil)
	}

	// filtering users urls
	urlRows := make(map[string]*models.URLRow)
	for _, shortURL := range shortURLs {
		//idx := slices.IndexFunc(found, func(u *models.URLRow) bool {
		//	return u.ShortURL == shortURL
		//})
		idx := -1
		for i := range found {
			if found[i].ShortURL == shortURL {
				idx = i
				break
			}
		}
		if idx > 0 {
			f := found[idx]
			urlRows[f.OrigURL] = f
		}
	}

	return checkUserURLs(userID, urlRows)
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
			err = e
			logger.Log.Debug("reading urls from file", zap.Error(err))
			break
		}

		d.urls[row.OrigURL] = row
		d.hash[row.ShortURL] = row

		_, ex := d.owners[row.UserID]
		if !ex {
			d.owners[row.UserID] = make([]*models.URLRow, 0)
		}
		d.owners[row.UserID] = append(d.owners[row.UserID], row)

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
