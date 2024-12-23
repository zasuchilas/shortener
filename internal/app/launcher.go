// Package app builds, starts and stops the service.
package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/secure"
	"github.com/zasuchilas/shortener/internal/app/server"
	"github.com/zasuchilas/shortener/internal/app/storage"
)

// App contains the application components.
type App struct {
	AppName             string
	AppVersion          string
	StorageInstanceName string
	ctx                 context.Context
	srv                 *server.Server
	store               storage.IStorage
	secure              *secure.Secure
}

// New creates the application instance.
func New(buildVersion, buildDate, buildCommit string) *App {
	// build info stdout
	log.Printf("Build version: %s \n", buildVersion)
	log.Printf("Build date: %s \n", buildDate)
	log.Printf("Build commit: %s \n", buildCommit)

	config.ParseFlags()
	ctx := context.Background()

	return &App{
		AppName:    "shortener",
		AppVersion: buildVersion,
		ctx:        ctx,
	}
}

// Run launches the application.
func (a *App) Run() {
	// logger
	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Fatal(err.Error())
	}
	logger.ServiceInfo(a.AppVersion)

	// storage
	if config.DatabaseDSN != "" {
		a.store = storage.NewDBPgsql()
	} else if config.FileStoragePath != "" {
		a.store = storage.NewDBFile()
	} else {
		a.store = storage.NewDBMaps()
	}
	a.StorageInstanceName = a.store.InstanceName()

	// secure service
	a.secure = secure.New(config.SecretKey, a.StorageInstanceName, config.SecureFilePath)

	// http server
	a.srv = server.New(a.store, a.secure)
	go a.srv.Start()

	// graceful shutdown
	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigint
		logger.Log.Info("The stop signal has been received", zap.String("signal", sig.String()))
		a.srv.Stop()
		close(idleConnsClosed)
	}()
	// blocked until the stop signal
	<-idleConnsClosed
	// stopping services
	a.store.Stop()
	// fin.
	logger.Log.Info("URL shortening service stopped")
}
