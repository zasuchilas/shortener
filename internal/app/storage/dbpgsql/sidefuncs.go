package dbpgsql

import (
	"context"
	"database/sql"
	"github.com/zasuchilas/shortener/internal/app/logger"
	. "github.com/zasuchilas/shortener/internal/app/models"
	"go.uber.org/zap"
	"time"
)

func createTablesIfNeed(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	q := `create table if not exists urls (
					uuid serial primary key,
					short varchar(254) not null,
					original varchar(254) not null
				);`

	// TODO: use scheme
	// TODO: INDEX idx_urls_short & INDEX idx_urls_original

	_, err := db.ExecContext(ctx, q, nil)
	if err != nil {
		logger.Log.Fatal("creating postgresql tables", zap.Error(err))
	}
}

func findOrig(ctx context.Context, db *sql.DB, origURL string) (urlRow *URLRow, ok bool, err error) {
	err = db.QueryRowContext(ctx,
		"SELECT uuid, original, short FROM urls WHERE original = ?",
		origURL).Scan(urlRow.Uuid, urlRow.ShortURL, urlRow.OrigURL)
	switch {
	case err == sql.ErrNoRows:
		return nil, false, nil
	case err != nil:
		return nil, false, err
	default:
		return urlRow, true, nil
	}
}

func findShort(ctx context.Context, db *sql.DB, shortURL string) (urlRow *URLRow, ok bool, err error) {
	err = db.QueryRowContext(ctx,
		"SELECT uuid, original, short FROM urls WHERE short = ?",
		shortURL).Scan(urlRow.Uuid, urlRow.ShortURL, urlRow.OrigURL)
	switch {
	case err == sql.ErrNoRows:
		return nil, false, nil
	case err != nil:
		return nil, false, err
	default:
		return urlRow, true, nil
	}
}

func (d *DBPgsql) isExist(shortURL string) (bool, error) {
	var uuid int64
	err := d.db.QueryRowContext(context.Background(),
		"SELECT uuid FROM urls WHERE short = ?",
		shortURL).Scan(&uuid)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func writeRow(ctx context.Context, db *sql.DB, shortURL, origURL string) error {
	result, err := db.ExecContext(ctx,
		"INSERT INTO urls (short, original) VALUES (?, ?)",
		shortURL, origURL)
	if err != nil {
		return err
	}
	uuid, err := result.LastInsertId()
	if err != nil {
		return err
	}
	logger.Log.Debug("row inserted into postgresql", zap.Int64("uuid", uuid))
	return nil
}
