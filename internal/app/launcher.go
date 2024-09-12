package app

import (
	"context"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/server"
	"github.com/zasuchilas/shortener/internal/app/storage"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	logLevel = "info"
)

type App struct {
	AppName    string
	AppVersion string
	ctx        context.Context
	waitGroup  *sync.WaitGroup
	server     server.Server
	store      *storage.Database
}

func New() *App {
	config.ParseFlags()
	ctx := context.Background()
	waitGroup := &sync.WaitGroup{}

	return &App{
		AppName:    "shortener",
		AppVersion: "0.0.0",
		ctx:        ctx,
		waitGroup:  waitGroup,
	}
}

func (a *App) Run() {
	if err := logger.Initialize(logLevel); err != nil {
		log.Fatal(err.Error())
	}
	logger.ServiceInfo(a.AppVersion)

	a.store = storage.New()

	srv := server.New(a.store)
	a.waitGroup.Add(1)
	go srv.Start()

	a.shutdown()
	a.waitGroup.Wait()
}

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
