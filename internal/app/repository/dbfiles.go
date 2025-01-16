package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/zasuchilas/shortener/internal/app/model"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/utils/filefuncs"
	"github.com/zasuchilas/shortener/internal/app/utils/hashfuncs"
)

var (
	_ IStorage = (*DBFiles)(nil)
)

// DBFiles is a file storage implementation.
type DBFiles struct {
	urls     map[string]*model.URLRow
	hash     map[string]*model.URLRow
	owners   map[int64][]*model.URLRow
	original []*model.URLRow
	lastID   int64
	mutex    sync.RWMutex
}

// NewDBFile creates an instance of the component.
func NewDBFile() *DBFiles {
	db := &DBFiles{
		urls:   make(map[string]*model.URLRow),
		hash:   make(map[string]*model.URLRow),
		owners: make(map[int64][]*model.URLRow),
		mutex:  sync.RWMutex{},
	}

	lastID, err := db.loadFromFile()
	if err != nil {
		logger.Log.Fatal("loading data from file", zap.Error(err))
	}
	db.lastID = lastID

	return db
}

// Stop stops the component.
func (d *DBFiles) Stop() {}

// InstanceName returns current instance name.
func (d *DBFiles) InstanceName() string {
	return InstanceFile
}

// WriteURL writes URL in the storage.
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

// ReadURL reads URL from the storage.
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

// Ping pings the storage.
//
// Not applicable for file storage instance.
func (d *DBFiles) Ping(_ context.Context) error {
	return errors.New("not allowed")
}

// WriteURLs writes URLs in the storage.
func (d *DBFiles) WriteURLs(ctx context.Context, origURLs []string, userID int64) (urlRows map[string]*model.URLRow, err error) {

	urlRows = make(map[string]*model.URLRow)

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
		d.owners[userID] = make([]*model.URLRow, 0)
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
			nextURLRow := &model.URLRow{
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
			d.original = append(d.original, nextURLRow)
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

// UserURLs returns user URLs from storage.
func (d *DBFiles) UserURLs(_ context.Context, userID int64) (urlRowList []*model.URLRow, err error) {

	d.mutex.RLock()
	found, ok := d.owners[userID]
	d.mutex.RUnlock()

	if !ok || len(found) == 0 {
		return nil, fmt.Errorf("%w", ErrNotFound)
	}

	return found, nil
}

// CheckDeletedURLs checks deleting URLs.
func (d *DBFiles) CheckDeletedURLs(_ context.Context, userID int64, shortURLs []string) error {
	// getting urls from request
	urlRows := make(map[string]*model.URLRow)
	d.mutex.RLock()
	for _, shortURL := range shortURLs {
		found, ok := d.hash[shortURL]
		if !ok {
			continue
		}
		urlRows[shortURL] = found
	}
	d.mutex.RUnlock()

	return checkUserURLs(userID, urlRows)
}

// DeleteURLs deletes URLs from the storage.
func (d *DBFiles) DeleteURLs(_ context.Context, shortURLs ...string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	for _, shortURL := range shortURLs {

		found, ok := d.hash[shortURL]
		if !ok {
			continue
		}

		found.Deleted = true
		// since found is a pointer, the value must change in all components
		// (url, hash, owner and original)
	}

	// rewrite file storage from original component
	w, err := filefuncs.NewFileReWriter(config.FileStoragePath)
	if err != nil {
		return err
	}
	defer w.Close()

	for _, line := range d.original {
		logger.Log.Debug("lines", zap.Any("line", line))
		err = w.WriteURLRow(line)
		if err != nil {
			break
		}
	}
	if err != nil {
		return err
	}

	return nil
}

// Stats returns count of URLs.
func (d *DBFiles) Stats(_ context.Context) (int, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return len(d.urls), nil
}

// TODO: as an option: use cache lib with reading from file

// loadFromFile loads URLs from the storage.
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
			d.owners[row.UserID] = make([]*model.URLRow, 0)
		}
		d.owners[row.UserID] = append(d.owners[row.UserID], row)

		d.original = append(d.original, row)

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
