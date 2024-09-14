package dbpgsql

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/storage/hashfuncs"
	"go.uber.org/zap"
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

	logger.Log.Debug("find in postgresql storage", zap.String("origURL", origURL))
	v, found, err := findOrig(ctx, d.db, origURL)
	if err != nil {
		logger.Log.Error("finding original url in postgresql storage", zap.Error(err), zap.String("origURL", origURL))
		return "", false, err
	}
	if found {
		return v.ShortURL, true, nil
	}

	// TODO: other way better

	logger.Log.Debug("making short url", zap.String("origURL", origURL))
	shortURL, err = hashfuncs.MakeShortURL(d.isExist)
	if err != nil {
		return "", false, err
	}

	logger.Log.Debug("append to postgresql storage", zap.String("origURL", origURL))
	err = writeRow(ctx, d.db, shortURL, origURL)
	if err != nil {
		return "", false, err
	}

	return shortURL, false, nil
}

func (d *DBPgsql) ReadURL(ctx context.Context, shortURL string) (origURL string, err error) {
	v, found, err := findShort(ctx, d.db, shortURL)
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
