package dbpgsql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
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
	createTablesIfNeed(pg)

	return &DBPgsql{pg}
}

func (d *DBPgsql) Stop() {
	if d.db != nil {
		_ = d.db.Close()
	}
}

func (d *DBPgsql) WriteURL(ctx context.Context, origURL string) (shortURL string, was bool, err error) {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	logger.Log.Debug("start tx in postgresql storage", zap.String("origURL", origURL))
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()

	// INSERT INTO urls (uuid, short, original) VALUES ($1, $2, $3) ... so SERIAL will break
	// INSERT INTO urls (short, original) VALUES ($1, $2) ON CONFLICT DO NOTHING ... so urls_uuid_seq will break
	// IT IS NECESSARY: if origURL already exists, do nothing (including not changing the urls_uuid_set counter)
	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urls (short, original) "+
			"SELECT $1, $2 "+
			"WHERE NOT EXISTS (SELECT 1 FROM urls WHERE original = $3) "+
			"RETURNING uuid")
	if err != nil {
		logger.Log.Error("preparing stmt", zap.Error(err))
		return "", false, err
	}
	defer stmt.Close()

	select {
	case <-ctxTm.Done():
		err = fmt.Errorf("the operation was canceled")
	default:
		err = nil
	}
	if err != nil {
		return "", false, err
	}

	logger.Log.Debug("getting last/next uuid from urls_uuid_seq")
	nextUUID, err := getNextUUID(ctx, d.db)
	logger.Log.Debug("next uuid", zap.Int64("uuid", nextUUID))
	if err != nil {
		logger.Log.Error("getting next uuid", zap.Error(err))
		return "", false, err
	}
	shortURL = hashfuncs.EncodeZeroHash(nextUUID)

	_, err = stmt.ExecContext(ctx, shortURL, origURL, origURL)
	if err != nil {
		logger.Log.Error("executing stmt", zap.Error(err))
		return "", false, err
	}

	logger.Log.Debug("closing tx")
	err = tx.Commit()
	if err != nil {
		logger.Log.Debug("closing tx exception (if we get this error tx will rollback ?)", zap.Error(err))
		return "", false, err
	}

	logger.Log.Debug("getting inserted url")
	v, _, err := findByOrig(ctx, d.db, origURL)
	if err != nil {
		logger.Log.Error("finding inserted url in postgresql storage (not impossible)", zap.Error(err), zap.String("origURL", origURL))
		return "", false, err
	}
	return v.ShortURL, true, nil
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
