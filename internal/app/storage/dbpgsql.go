package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"github.com/zasuchilas/shortener/internal/app/utils/hashfuncs"
)

var (
	_ IStorage = (*DBPgsql)(nil)
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

	db := &DBPgsql{
		db: pg,
	}

	return db
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
	found, ex, err := findByOrig(ctx, d.db, origURL)
	if err != nil {
		return "", false, err
	}
	if ex {
		return found.ShortURL, true, nil
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
	found, ex, err := findByShort(ctx, d.db, shortURL)
	if err != nil {
		return "", err
	}

	if !ex {
		return "", fmt.Errorf("%w", ErrNotFound)
	}

	if found.Deleted {
		return "", fmt.Errorf("%w", ErrGone)
	}

	return found.OrigURL, nil
}

func (d *DBPgsql) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *DBPgsql) WriteURLs(
	ctx context.Context,
	origURLs []string,
	userID int64,
) (urlRows map[string]*models.URLRow, err error) {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	logger.Log.Debug("start tx in postgresql storage")
	tx, err := d.db.BeginTx(ctxTm, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// INSERT INTO urls (uuid, short, original) VALUES ($1, $2, $3) ... so SERIAL will break
	// INSERT INTO urls (short, original) VALUES ($1, $2) ON CONFLICT DO NOTHING ... so urls_uuid_seq will break
	// IT IS NECESSARY: if origURL already exists, do nothing (including not changing the urls_uuid_set counter)
	stmt, err := tx.PrepareContext(ctxTm,
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
			logger.Log.Info("got next shortURL candidate & next id",
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

func (d *DBPgsql) UserURLs(ctx context.Context, userID int64) (urlRowList []*models.URLRow, err error) {
	found, ex, err := findByUser(ctx, d.db, userID)
	if !ex {
		return nil, fmt.Errorf("%w", ErrNotFound)
	}
	if err != nil {
		return nil, err
	}

	return found, nil
}

func (d *DBPgsql) CheckDeletedURLs(ctx context.Context, userID int64, shortURLs []string) error {
	urlRows, err := selectByShortURLs(ctx, d.db, shortURLs)
	if err != nil {
		return err
	}
	return checkUserURLs(userID, urlRows)
}

func (d *DBPgsql) DeleteURLs(ctx context.Context, shortURLs ...string) error {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctxTm, `UPDATE urls SET deleted = true WHERE short = any($1)`)
	if err != nil {
		logger.Log.Info("preparing stmt", zap.String("error", err.Error()))
		return err
	}
	defer stmt.Close()

	select {
	case <-ctxTm.Done():
		return fmt.Errorf("the operation was canceled")
	default:
		_, err = stmt.ExecContext(ctxTm, shortURLs)
		if err != nil {
			return err
		}
		logger.Log.Info("urls deleted", zap.String("shortURLs", strings.Join(shortURLs, ", ")))
	}

	return nil
}

func createTablesIfNeed(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	q := `CREATE TABLE IF NOT EXISTS urls (
					id SERIAL PRIMARY KEY,
					short VARCHAR(254) NOT NULL,
					original VARCHAR(254) NOT NULL UNIQUE,
    			user_id INTEGER NOT NULL DEFAULT 0,
    			deleted BOOL NOT NULL DEFAULT false
				);
				CREATE INDEX IF NOT EXISTS idx_short ON urls (short);
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
		logger.Log.Error("creating query", zap.String("error", err.Error()))
		return nil, err
	}
	defer rows.Close()

	urlRows = make(map[string]*models.URLRow)
	for rows.Next() {
		var urlRow models.URLRow
		err = rows.Scan(&urlRow.ID, &urlRow.ShortURL, &urlRow.OrigURL)
		if err != nil {
			logger.Log.Error("scanning rows", zap.String("error", err.Error()))
			return nil, err
		}
		urlRows[urlRow.OrigURL] = &urlRow
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("checkin rows on errors", zap.String("error", err.Error()))
		return nil, err
	}

	return urlRows, nil
}

func selectByShortURLs(ctx context.Context, db *sql.DB, shortURLs []string) (urlRows map[string]*models.URLRow, err error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, short, original, user_id, deleted FROM urls WHERE short = any($1)`,
		shortURLs)
	if err != nil {
		logger.Log.Error("creating query", zap.String("error", err.Error()))
		return nil, err
	}
	defer rows.Close()

	urlRows = make(map[string]*models.URLRow)
	for rows.Next() {
		var urlRow models.URLRow
		err = rows.Scan(&urlRow.ID, &urlRow.ShortURL, &urlRow.OrigURL, &urlRow.UserID, &urlRow.Deleted)
		if err != nil {
			logger.Log.Error("scanning rows", zap.String("error", err.Error()))
			return nil, err
		}
		urlRows[urlRow.OrigURL] = &urlRow
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("checkin rows on errors", zap.String("error", err.Error()))
		return nil, err
	}

	return urlRows, nil
}

func findByShort(ctx context.Context, db *sql.DB, shortURL string) (urlRow *models.URLRow, exist bool, err error) {
	var v models.URLRow
	err = db.QueryRowContext(ctx,
		"SELECT id, short, original, user_id, deleted FROM urls WHERE short = $1",
		shortURL).Scan(&v.ID, &v.ShortURL, &v.OrigURL, &v.UserID, &v.Deleted)
	switch {
	case err == sql.ErrNoRows:
		return nil, false, nil
	case err != nil:
		return nil, false, err
	default:
		return &v, true, nil
	}
}

func findByOrig(ctx context.Context, db *sql.DB, origURL string) (urlRow *models.URLRow, exist bool, err error) {
	var v models.URLRow
	err = db.QueryRowContext(ctx,
		"SELECT id, short, original, user_id FROM urls WHERE original = $1",
		origURL).Scan(&v.ID, &v.ShortURL, &v.OrigURL, &v.UserID)
	switch {
	case err == sql.ErrNoRows:
		return nil, false, nil
	case err != nil:
		return nil, false, err
	default:
		return &v, true, nil
	}
}

func findByUser(ctx context.Context, db *sql.DB, userID int64) (urlRowList []*models.URLRow, exist bool, err error) {

	rows, err := db.QueryContext(ctx,
		"SELECT id, short, original, user_id FROM urls WHERE user_id = $1",
		userID)
	if err != nil {
		logger.Log.Error("creating query", zap.String("error", err.Error()))
		return nil, false, err
	}
	defer rows.Close()

	urlRowList = make([]*models.URLRow, 0)
	for rows.Next() {
		var v models.URLRow
		err = rows.Scan(&v.ID, &v.ShortURL, &v.OrigURL, &v.UserID)
		if err != nil {
			logger.Log.Error("scanning rows", zap.String("error", err.Error()))
			return nil, false, err
		}
		urlRowList = append(urlRowList, &v)
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("checkin rows on errors", zap.String("error", err.Error()))
		return nil, false, err
	}

	if len(urlRowList) == 0 {
		return nil, false, errors.New("not found")
	}

	return urlRowList, true, nil
}
