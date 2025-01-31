package converter

import (
	"github.com/zasuchilas/shortener/internal/app/model"
	"github.com/zasuchilas/shortener/pkg/shortenergrpcv1"
	"github.com/zasuchilas/shortener/pkg/shortenerhttpv1"
)

// ToHTTPFromStats _
func ToHTTPFromStats(in model.Stats) shortenerhttpv1.StatsResponse {
	return shortenerhttpv1.StatsResponse{
		URLs:  in.URLs,
		Users: in.Users,
	}
}

// ToGRPCFromStats _
func ToGRPCFromStats(in *model.Stats) *shortenergrpcv1.StatsResponse {
	return &shortenergrpcv1.StatsResponse{
		Urls:  int64(in.URLs),
		Users: int64(in.Users),
	}
}
