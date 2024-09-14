package dbfiles

import (
	"errors"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"go.uber.org/zap"
	"io"
	"time"
)

// TODO: cache (use lib)

func (d *DBFiles) isExist(shortURL string) (bool, error) {
	_, found, err := d.lookup(nil, nil, &shortURL)
	if err != nil {
		return false, err
	}
	return found, nil
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

func (d *DBFiles) writeRow(shortURL, origURL string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	w, err := newFileWriter(config.FileStoragePath)
	if err != nil {
		return err
	}
	defer w.close()

	return w.writeURLRow(time.Now().Unix(), shortURL, origURL)
}
