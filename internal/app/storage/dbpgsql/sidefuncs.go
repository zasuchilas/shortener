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
					short varchar(254) not null UNIQUE,
					original varchar(254) not null UNIQUE
				);`

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		logger.Log.Fatal("creating postgresql tables", zap.Error(err))
	}
}

func getNextUUID(ctx context.Context, tx *sql.Tx) (int64, error) {
	var lastUUID int64
	var isCalled bool
	err := tx.QueryRowContext(ctx,
		`SELECT last_value, is_called FROM urls_uuid_seq`).Scan(&lastUUID, &isCalled)
	if err != nil {
		return 0, err
	}
	if !isCalled {
		return lastUUID, nil
	}
	return lastUUID + 1, nil
}

func selectByOrigURLs(ctx context.Context, db *sql.DB, origURLs []string) (urlRows map[string]*models.URLRow, err error) {

	logger.Log.Debug("selectByOrigURLs", zap.Any("origURLs", origURLs))
	rows, err := db.QueryContext(ctx,
		`SELECT uuid, short, original FROM urls WHERE original = any($1)`,
		origURLs) // strings.Join(origURLs, ","))
	if err != nil {
		logger.Log.Error("creating query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	urlRows = make(map[string]*models.URLRow)
	for rows.Next() {
		logger.Log.Debug("row")
		var urlRow models.URLRow
		err = rows.Scan(&urlRow.UUID, &urlRow.ShortURL, &urlRow.OrigURL)
		if err != nil {
			logger.Log.Error("scanning rows", zap.Error(err))
			return nil, err
		}
		logger.Log.Debug("row next", zap.Any("row", urlRow))
		urlRows[urlRow.OrigURL] = &urlRow
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("checkin rows on errors", zap.Error(err))
		return nil, err
	}

	return urlRows, nil
}

func findByShort(ctx context.Context, db *sql.DB, shortURL string) (urlRow *models.URLRow, found bool, err error) {
	var v models.URLRow
	err = db.QueryRowContext(ctx,
		"SELECT uuid, short, original FROM urls WHERE short = $1",
		shortURL).Scan(&v.UUID, &v.ShortURL, &v.OrigURL)
	switch {
	case err == sql.ErrNoRows:
		return nil, false, nil
	case err != nil:
		return nil, false, err
	default:
		return &v, true, nil
	}
}

func findByOrig(ctx context.Context, db *sql.DB, origURL string) (urlRow *models.URLRow, found bool, err error) {
	var v models.URLRow
	err = db.QueryRowContext(ctx,
		"SELECT uuid, short, original FROM urls WHERE original = $1",
		origURL).Scan(&v.UUID, &v.ShortURL, &v.OrigURL)
	switch {
	case err == sql.ErrNoRows:
		return nil, false, nil
	case err != nil:
		return nil, false, err
	default:
		return &v, true, nil
	}
}
