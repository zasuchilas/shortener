package main

import (
	"github.com/zasuchilas/shortener/internal/app"
)

func main() {
	service := app.New()
	service.Run()
}
