package app

import (
	"context"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/server"
	"github.com/zasuchilas/shortener/internal/app/storage"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	addr = "localhost:8080"
)

type App struct {
	AppName    string
	AppVersion string
	ctx        context.Context
	waitGroup  *sync.WaitGroup
	server     server.Server
}

func New() *App {
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
	logger.ServiceInfo(a.AppVersion)

	st := storage.New() // TODO: use or not interface Storage here ?

	srv := server.New(addr, st)
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
		log.Printf("The %s stop signal has been received", sig)
		close(sigChan)

		// TODO: stop app components

		// TODO: stop server

		log.Println("URL shortening service stopped")
		a.waitGroup.Done()
	}()

}
