package dbpgsql

import (
	"context"
	"database/sql"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"go.uber.org/zap"
	"strings"
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
	if isCalled == false {
		return lastUUID, nil
	}
	return lastUUID + 1, nil
}

func selectByOrigURLs(ctx context.Context, db *sql.DB, origURLs []string) (shortURLs []string, err error) {

	rows, err := db.QueryContext(ctx,
		`SELECT short FROM urls WHERE original IN ($1)`,
		strings.Join(origURLs, ","))
	if err != nil {
		logger.Log.Error("creating query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var shortURL string
		err = rows.Scan(&shortURL)
		if err != nil {
			logger.Log.Error("scanning rows", zap.Error(err))
			return nil, err
		}
		shortURLs = append(shortURLs, shortURL)
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("checkin rows on errors", zap.Error(err))
		return nil, err
	}

	return shortURLs, nil
}

func findByShort(ctx context.Context, db *sql.DB, shortURL string) (urlRow *models.URLRow, ok bool, err error) {
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
