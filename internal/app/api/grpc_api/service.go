package grpc_api

import (
	"github.com/zasuchilas/shortener/internal/app/service"
	desc "github.com/zasuchilas/shortener/pkg/shortener_v1"
)

type Implementation struct {
	desc.UnimplementedShortenerV1Server
	shortenerService service.ShortenerService
}

func NewImplementation(shortenerService service.ShortenerService) *Implementation {
	return &Implementation{shortenerService: shortenerService}
}
