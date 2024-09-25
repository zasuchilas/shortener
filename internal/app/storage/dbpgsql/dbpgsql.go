package dbpgsql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"github.com/zasuchilas/shortener/internal/app/storage/hashfuncs"
	"go.uber.org/zap"
	"time"
)

// DBPgsql is a postgresql storage implementation
type DBPgsql struct {
	db *sql.DB
}

func New() *DBPgsql {
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

func (d *DBPgsql) WriteURL(ctx context.Context, origURL string) (shortURL string, conflict bool, err error) {

	logger.Log.Debug("checking if already exist")
	row, found, err := findByOrig(ctx, d.db, origURL)
	if err != nil {
		return "", false, err
	}
	if found {
		return row.ShortURL, true, nil
	}

	logger.Log.Debug("writing URL")
	urlRows, err := d.WriteURLs(ctx, []string{origURL})
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

func (d *DBPgsql) WriteURLs(ctx context.Context, origURLs []string) (urlRows map[string]*models.URLRow, err error) {

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
		"INSERT INTO urls (short, original) "+
			"SELECT $1, $2 "+
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
			logger.Log.Debug("getting last/next uuid from urls_uuid_seq")
			var nextUUID int64
			nextUUID, err = getNextUUID(ctx, tx)
			if err != nil {
				logger.Log.Error("getting next uuid", zap.Error(err))
				break loop
			}
			shortURLCandidate := hashfuncs.EncodeZeroHash(nextUUID)
			logger.Log.Debug("got next shortURL candidate & next uuid",
				zap.String("shortURLCandidate", shortURLCandidate), zap.Int64("uuid", nextUUID))

			logger.Log.Debug("executing stmt")
			_, err = stmt.ExecContext(ctx, shortURLCandidate, origURL, origURL)
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

func (d *DBPgsql) NewUser(_ context.Context) (userID int64, err error) {
	return 0, err
}
