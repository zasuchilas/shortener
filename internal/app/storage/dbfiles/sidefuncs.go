package dbfiles

import (
	"errors"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"github.com/zasuchilas/shortener/internal/app/storage/hashfuncs"
	"go.uber.org/zap"
	"io"
)

// as an option: use cache lib with reading from file

func (d *DBFiles) isExist(shortURL string) (bool, error) {
	_, found, err := d.lookup(nil, nil, &shortURL)
	if err != nil {
		return false, err
	}
	return found, nil
}

func (d *DBFiles) loadFromFile() (lastID int64, err error) {
	r, err := newFileReader(config.FileStoragePath)
	if err != nil {
		return 0, err
	}
	defer r.close()

	var lastHash string
	for {
		row, e := r.readURLRow()
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

func (d *DBFiles) lookup(uuid *int64, origURL, shortURL *string) (urlRow *models.URLRow, ok bool, err error) {
	condCount := 0
	if uuid != nil {
		condCount++
	}
	if origURL != nil {
		condCount++
	}
	if shortURL != nil {
		condCount++
	}
	if condCount == 0 {
		return nil, false, errors.New("empty conditions for lookup url row in file storage")
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	r, err := newFileReader(config.FileStoragePath)
	if err != nil {
		return nil, false, err
	}
	defer r.close()

	for {
		row, e := r.readURLRow()
		if e == io.EOF {
			break
		}
		if e != nil {
			logger.Log.Debug("reading url from file", zap.Error(err))
			err = e
			break
		}

		cc := 0
		if uuid != nil && row.UUID == *uuid {
			cc++
		}
		if origURL != nil && row.OrigURL == *origURL {
			cc++
		}
		if shortURL != nil && row.ShortURL == *shortURL {
			cc++
		}

		if cc == condCount {
			urlRow = row
			ok = true
			break
		}
	}
	return urlRow, ok, err
}

func (d *DBFiles) writeRow(uuid int64, shortURL, origURL string) error {
	w, err := newFileWriter(config.FileStoragePath)
	if err != nil {
		return err
	}
	defer w.close()

	return w.writeURLRow(uuid, shortURL, origURL)
}
