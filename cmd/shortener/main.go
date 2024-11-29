package main

import (
	_ "net/http/pprof"

	"github.com/zasuchilas/shortener/internal/app"
)

func main() {
	service := app.New()
	service.Run()
}
