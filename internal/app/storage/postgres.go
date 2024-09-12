package storage

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"go.uber.org/zap"
)

func NewPostgres() *sql.DB {
	db, err := sql.Open("pgx", config.DatabaseDSN)
	if err != nil {
		logger.Log.Fatal("opening connection to postgresql", zap.Error(err))
		return nil
	}
	return db
}
