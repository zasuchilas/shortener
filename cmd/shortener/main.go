package main

import (
	"github.com/zasuchilas/shortener/internal/app"
	_ "net/http/pprof"
)

func main() {
	service := app.New()
	service.Run()
}
