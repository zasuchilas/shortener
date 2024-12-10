// Package app builds, starts and stops the service.
package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
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
	waitGroup           *sync.WaitGroup
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
	waitGroup := &sync.WaitGroup{}

	return &App{
		AppName:    "shortener",
		AppVersion: buildVersion,
		ctx:        ctx,
		waitGroup:  waitGroup,
	}
}

// Run launches the application.
func (a *App) Run() {
	if err := logger.Initialize(config.LogLevel); err != nil {
		log.Fatal(err.Error())
	}

	logger.ServiceInfo(a.AppVersion)
	logger.ConfigInfo()

	if config.DatabaseDSN != "" {
		a.store = storage.NewDBPgsql()
	} else if config.FileStoragePath != "" {
		a.store = storage.NewDBFile()
	} else {
		a.store = storage.NewDBMaps()
	}
	a.StorageInstanceName = a.store.InstanceName()

	a.secure = secure.New(config.SecretKey, a.StorageInstanceName, config.SecureFilePath)

	a.srv = server.New(a.store, a.secure)
	a.waitGroup.Add(1)
	go a.srv.Start()

	a.shutdown()
	a.waitGroup.Wait()
}

// shutdown intercepts exit signals and performs a graceful shutdown.
func (a *App) shutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		logger.Log.Info("The stop signal has been received", zap.String("signal", sig.String()))
		close(sigChan)

		// TODO: stop app components

		// TODO: stop server
		a.store.Stop()

		logger.Log.Info("URL shortening service stopped")
		a.waitGroup.Done()
	}()
}
