package main

import (
	_ "net/http/pprof"

	"github.com/zasuchilas/shortener/internal/app"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	service := app.New(buildVersion, buildDate, buildCommit)
	service.Run()
}
