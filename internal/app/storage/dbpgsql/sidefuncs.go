package dbpgsql

import (
	"context"
	"database/sql"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"go.uber.org/zap"
	"time"
)

func createTablesIfNeed(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	q := `CREATE TABLE IF NOT EXISTS urls (
					uuid serial primary key,
					short varchar(254) not null,
					original varchar(254) not null
				);`

	// TODO: use scheme
	// TODO: INDEX idx_urls_short & INDEX idx_urls_original

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		logger.Log.Fatal("creating postgresql tables", zap.Error(err))
	}
}

func findByOrig(ctx context.Context, db *sql.DB, origURL string) (urlRow *models.URLRow, ok bool, err error) {
	var v models.URLRow
	err = db.QueryRowContext(ctx,
		`SELECT uuid, short, original FROM urls WHERE original = $1`,
		origURL).Scan(&v.Uuid, &v.ShortURL, &v.OrigURL)
	switch {
	case err == sql.ErrNoRows:
		return nil, false, nil
	case err != nil:
		return nil, false, err
	default:
		return &v, true, nil
	}
}

func findByShort(ctx context.Context, db *sql.DB, shortURL string) (urlRow *models.URLRow, ok bool, err error) {
	var v models.URLRow
	err = db.QueryRowContext(ctx,
		"SELECT uuid, short, original FROM urls WHERE short = $1",
		shortURL).Scan(&v.Uuid, &v.ShortURL, &v.OrigURL)
	switch {
	case err == sql.ErrNoRows:
		return nil, false, nil
	case err != nil:
		return nil, false, err
	default:
		return &v, true, nil
	}
}

func (d *DBPgsql) isExist(shortURL string) (bool, error) {
	var uuid int64
	err := d.db.QueryRowContext(context.Background(),
		"SELECT uuid FROM urls WHERE short = $1",
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
	_, err := db.ExecContext(ctx,
		"INSERT INTO urls (short, original) VALUES ($1, $2)",
		shortURL, origURL)
	if err != nil {
		return err
	}
	//uuid, err := result.LastInsertId() // TODO: LastInsertId is not supported by this driver
	//if err != nil {
	//	return err
	//}
	//logger.Log.Debug("row inserted into postgresql", zap.Int64("uuid", uuid))
	logger.Log.Debug("row inserted into postgresql")
	return nil
}
