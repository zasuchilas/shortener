package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"github.com/zasuchilas/shortener/internal/app/utils/hashfuncs"
)

var (
	_ IStorage = (*DBMaps)(nil)
)

// DBMaps is a RAM storage on double maps.
type DBMaps struct {
	urls   map[string]*models.URLRow
	hash   map[string]*models.URLRow
	owners map[int64][]*models.URLRow
	lastID int64
	mutex  sync.RWMutex
}

// NewDBMaps creates an instance of the component.
func NewDBMaps() *DBMaps {
	db := &DBMaps{
		urls:   make(map[string]*models.URLRow),
		hash:   make(map[string]*models.URLRow),
		owners: make(map[int64][]*models.URLRow),
	}
	return db
}

// Stop stops the component.
func (d *DBMaps) Stop() {}

// InstanceName returns current instance name.
func (d *DBMaps) InstanceName() string {
	return InstanceMemory
}

// WriteURL writes URL in the storage.
func (d *DBMaps) WriteURL(ctx context.Context, origURL string, userID int64) (shortURL string, conflict bool, err error) {
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
func (d *DBMaps) ReadURL(_ context.Context, shortURL string) (origURL string, err error) {
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
func (d *DBMaps) Ping(_ context.Context) error {
	return errors.New("not allowed")
}

// WriteURLs writes URLs in the storage.
func (d *DBMaps) WriteURLs(ctx context.Context, origURLs []string, userID int64) (urlRows map[string]*models.URLRow, err error) {

	urlRows = make(map[string]*models.URLRow)

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// start ~tx in maps storage
	d.mutex.Lock()
	defer d.mutex.Unlock()

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
			logger.Log.Debug("find is ready in maps storage", zap.String("origURL", origURL))
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

// UserURLs returns user URLs from storage.
func (d *DBMaps) UserURLs(_ context.Context, userID int64) (urlRowList []*models.URLRow, err error) {
	d.mutex.RLock()
	found, ok := d.owners[userID]
	d.mutex.RUnlock()

	if !ok || len(found) == 0 {
		return nil, fmt.Errorf("%w", ErrNotFound)
	}

	return found, nil
}

// CheckDeletedURLs checks deleting URLs.
func (d *DBMaps) CheckDeletedURLs(_ context.Context, userID int64, shortURLs []string) error {
	// getting urls from request
	urlRows := make(map[string]*models.URLRow)
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
func (d *DBMaps) DeleteURLs(_ context.Context, shortURLs ...string) error {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	for _, shortURL := range shortURLs {

		found, ok := d.hash[shortURL]
		if !ok {
			continue
		}

		found.Deleted = true // since found is a pointer, the value must change in all components (url, hash, owner)
	}

	return nil
}

// Stats returns count of URLs.
func (d *DBMaps) Stats(_ context.Context) (int, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return len(d.urls), nil
}

// Write is for testing usage
//func Write(st *DBMaps, id, userID int64, shortURL, origURL string) {
//	// for testing usage
//	//st.urls["http://спорт.ru/"] = "abcdefgh"
//	//st.hash["abcdefgh"] = "http://спорт.ru/"
//
//	urlRow := &models.URLRow{
//		ID:       id,
//		ShortURL: shortURL,
//		OrigURL:  origURL,
//		UserID:   userID,
//		Deleted:  false,
//	}
//	st.urls[origURL] = urlRow
//	st.hash[shortURL] = urlRow
//
//	_, ex := st.owners[userID]
//	if !ex {
//		st.owners[userID] = make([]*models.URLRow, 0)
//	}
//	st.owners[userID] = append(st.owners[userID], urlRow)
//}
