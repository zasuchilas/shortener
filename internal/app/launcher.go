// Package app builds, starts and stops the service.
package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/api/http_api"
	"github.com/zasuchilas/shortener/internal/app/grpc_server"
	"github.com/zasuchilas/shortener/internal/app/http_server"
	"github.com/zasuchilas/shortener/internal/app/repository"
	"github.com/zasuchilas/shortener/internal/app/service/shortener"

	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/secure"
)

// App contains the application components.
type App struct {
	AppName             string
	AppVersion          string
	StorageInstanceName string
	ctx                 context.Context
	secure              *secure.Secure
	httpServer          *http_server.Server
	grpcServer          *grpc_server.Server
	shortenerRepo       repository.IStorage
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

	// repository
	a.initRepository()

	// secure service
	a.secure = secure.New(config.SecretKey, a.StorageInstanceName, config.SecureFilePath)

	// shortener service
	shortenerService := shortener.NewService(a.shortenerRepo, a.secure)

	// http server
	a.httpServer = http_server.NewServer(http_api.NewImplementation(shortenerService), a.secure)
	go a.httpServer.Run()

	// grpc server
	a.grpcServer = grpc_server.NewServer(shortenerService)
	go a.grpcServer.Run()

	// graceful shutdown
	a.initGracefulShutdown()
}

func (a *App) initRepository() {
	if config.DatabaseDSN != "" {
		a.shortenerRepo = repository.NewDBPgsql()
	} else if config.FileStoragePath != "" {
		a.shortenerRepo = repository.NewDBFile()
	} else {
		a.shortenerRepo = repository.NewDBMaps()
	}
	a.StorageInstanceName = a.shortenerRepo.InstanceName()
}

func (a *App) initGracefulShutdown() {
	idleConnsClosed := make(chan struct{})
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigint
		logger.Log.Info("The stop signal has been received", zap.String("signal", sig.String()))
		a.httpServer.Stop()
		close(idleConnsClosed)
	}()
	// blocked until the stop signal
	<-idleConnsClosed
	// stopping services
	a.shortenerRepo.Stop()
	// fin.
	logger.Log.Info("URL shortening service stopped")
}
