package dbfiles

import (
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/storage/hashfuncs"
	"go.uber.org/zap"
	"io"
)

// as an option: use cache lib with reading from file

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

func (d *DBFiles) writeRow(uuid int64, shortURL, origURL string) error {
	w, err := newFileWriter(config.FileStoragePath)
	if err != nil {
		return err
	}
	defer w.close()

	return w.writeURLRow(uuid, shortURL, origURL)
}
