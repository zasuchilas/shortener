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

	logger.Log.Debug("find is ready in postgresql storage", zap.String("origURL", origURL))
	v, found, err := findByOrig(ctx, d.db, origURL)
	if err != nil {
		logger.Log.Error("finding original url in postgresql storage", zap.Error(err), zap.String("origURL", origURL))
		return "", false, err
	}
	if found {
		return v.ShortURL, true, nil
	}

	logger.Log.Debug("start tx in postgresql storage", zap.String("origURL", origURL))
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()

	//stmt, err := tx.PrepareContext(ctx,
	//	"INSERT INTO urls (uuid, short, original) VALUES ($1, $2, $3)")
	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urls (short, original) VALUES ($1, $2)")
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

	nextUUID, err := getNextUUID(ctx, d.db)
	logger.Log.Debug("next uuid", zap.Int64("uuid", nextUUID))
	if err != nil {
		logger.Log.Error("getting next uuid", zap.Error(err))
		return "", false, err
	}
	shortURL = hashfuncs.EncodeZeroHash(nextUUID)

	//_, err = stmt.ExecContext(ctx, nextUUID, shortURL, origURL)
	_, err = stmt.ExecContext(ctx, shortURL, origURL)
	if err != nil {
		logger.Log.Error("executing stmt", zap.Error(err))
		return "", false, err
	}

	err = tx.Commit()
	if err != nil {
		return "", false, err
	}

	return shortURL, false, nil
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
