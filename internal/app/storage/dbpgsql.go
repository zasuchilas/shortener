package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"github.com/zasuchilas/shortener/internal/app/utils/hashfuncs"
	"go.uber.org/zap"
	"time"
)

// DBPgsql is a postgresql storage implementation
type DBPgsql struct {
	db *sql.DB
}

func NewDBPgsql() *DBPgsql {
	pg, err := sql.Open("pgx", config.DatabaseDSN)
	if err != nil {
		logger.Log.Fatal("opening connection to postgresql", zap.Error(err))
		return nil
	}

	logger.Log.Debug("creating db tables if need")
	createTablesIfNeed(pg) // TODO: constant table creation rows can be migrated

	return &DBPgsql{pg}
}

func (d *DBPgsql) Stop() {
	if d.db != nil {
		_ = d.db.Close()
	}
}

func (d *DBPgsql) InstanceName() string {
	return InstancePostgresql
}

func (d *DBPgsql) WriteURL(ctx context.Context, origURL string, userID int64) (shortURL string, conflict bool, err error) {

	logger.Log.Debug("checking if already exist")
	row, found, err := findByOrig(ctx, d.db, origURL)
	if err != nil {
		return "", false, err
	}
	if found {
		return row.ShortURL, true, nil
	}

	logger.Log.Debug("writing URL")
	urlRows, err := d.WriteURLs(ctx, []string{origURL}, userID)
	if err != nil {
		return "", false, err
	}
	if urlRows == nil || urlRows[origURL] == nil {
		return "", false, errors.New("something wrong with writing URL")
	}
	return urlRows[origURL].ShortURL, false, nil
}

func (d *DBPgsql) ReadURL(ctx context.Context, shortURL string) (origURL string, err error) {
	v, found, err := findByShort(ctx, d.db, shortURL)
	if err != nil {
		return "", err
	}
	if !found {
		return "", errors.New("not found")
	}

	return v.OrigURL, nil
}

func (d *DBPgsql) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *DBPgsql) WriteURLs(ctx context.Context, origURLs []string, userID int64) (urlRows map[string]*models.URLRow, err error) {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	logger.Log.Debug("start tx in postgresql storage")
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// INSERT INTO urls (uuid, short, original) VALUES ($1, $2, $3) ... so SERIAL will break
	// INSERT INTO urls (short, original) VALUES ($1, $2) ON CONFLICT DO NOTHING ... so urls_uuid_seq will break
	// IT IS NECESSARY: if origURL already exists, do nothing (including not changing the urls_uuid_set counter)
	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urls (short, original, user_id) "+
			"SELECT $1, $2, $4 "+
			"WHERE NOT EXISTS (SELECT 1 FROM urls WHERE original = $3)")
	if err != nil {
		logger.Log.Error("preparing stmt", zap.Error(err))
		return nil, err
	}
	defer stmt.Close()

loop:
	for _, origURL := range origURLs {
		select {
		case <-ctxTm.Done():
			err = fmt.Errorf("the operation was canceled")
			break loop
		default:
			// getting last/next id from urls_id_seq
			var nextID int64
			nextID, err = getNextUUID(ctx, tx)
			if err != nil {
				logger.Log.Error("getting next id", zap.Error(err))
				break loop
			}
			shortURLCandidate := hashfuncs.EncodeZeroHash(nextID)
			logger.Log.Debug("got next shortURL candidate & next id",
				zap.String("shortURLCandidate", shortURLCandidate), zap.Int64("id", nextID))

			logger.Log.Debug("executing stmt")
			_, err = stmt.ExecContext(ctx, shortURLCandidate, origURL, origURL, userID)
			if err != nil {
				logger.Log.Error("executing stmt", zap.Error(err))
				break loop
			}
		}
	}
	if err != nil {
		return nil, err
	}

	logger.Log.Debug("closing tx")
	err = tx.Commit()
	if err != nil {
		logger.Log.Debug("closing tx exception (if we get this error tx will must rollback)", zap.Error(err))
		return nil, err
	}

	logger.Log.Debug("getting inserted urls")
	urlRows, err = selectByOrigURLs(ctx, d.db, origURLs)
	if err != nil {
		logger.Log.Error("finding inserted url in postgresql storage (not impossible)", zap.Error(err))
		return nil, err
	}
	return urlRows, nil
}

func createTablesIfNeed(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	q := `CREATE TABLE IF NOT EXISTS urls (
					id SERIAL PRIMARY KEY,
					short VARCHAR(254) NOT NULL UNIQUE,
					original VARCHAR(254) NOT NULL UNIQUE,
    			user_id INTEGER NOT NULL DEFAULT 0,
    			deleted BOOL NOT NULL DEFAULT false
				);
				CREATE INDEX IF NOT EXISTS idx_user_id ON urls (user_id);
				CREATE INDEX IF NOT EXISTS idx_deleted ON urls (deleted);
				`

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		logger.Log.Fatal("creating postgresql tables", zap.Error(err))
	}
}

func getNextUUID(ctx context.Context, tx *sql.Tx) (int64, error) {
	var lastID int64
	var isCalled bool
	err := tx.QueryRowContext(ctx,
		`SELECT last_value, is_called FROM urls_id_seq`).Scan(&lastID, &isCalled)
	if err != nil {
		return 0, err
	}
	if !isCalled {
		return lastID, nil
	}
	return lastID + 1, nil
}

func selectByOrigURLs(ctx context.Context, db *sql.DB, origURLs []string) (urlRows map[string]*models.URLRow, err error) {

	logger.Log.Debug("selectByOrigURLs", zap.Any("origURLs", origURLs))
	rows, err := db.QueryContext(ctx,
		`SELECT id, short, original FROM urls WHERE original = any($1)`,
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
		err = rows.Scan(&urlRow.ID, &urlRow.ShortURL, &urlRow.OrigURL)
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
		"SELECT id, short, original FROM urls WHERE short = $1",
		shortURL).Scan(&v.ID, &v.ShortURL, &v.OrigURL)
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
		"SELECT id, short, original FROM urls WHERE original = $1",
		origURL).Scan(&v.ID, &v.ShortURL, &v.OrigURL)
	switch {
	case err == sql.ErrNoRows:
		return nil, false, nil
	case err != nil:
		return nil, false, err
	default:
		return &v, true, nil
	}
}
