package grpcapi

import (
	"github.com/zasuchilas/shortener/internal/app/service"
	desc "github.com/zasuchilas/shortener/pkg/shortenergrpcv1"
)

type Implementation struct {
	desc.UnimplementedShortenerV1Server
	shortenerService service.ShortenerService
}

func NewImplementation(shortenerService service.ShortenerService) *Implementation {
	return &Implementation{shortenerService: shortenerService}
}
