package converter

import (
	"github.com/zasuchilas/shortener/internal/app/model"
	"github.com/zasuchilas/shortener/pkg/shortener_http_api_v1"
	"github.com/zasuchilas/shortener/pkg/shortener_v1"
)

func ToHTTPFromStats(in model.Stats) shortener_http_api_v1.StatsResponse {
	return shortener_http_api_v1.StatsResponse{
		URLs:  in.URLs,
		Users: in.Users,
	}
}

func ToGRPCFromStats(in *model.Stats) *shortener_v1.StatsResponse {
	return &shortener_v1.StatsResponse{
		Urls:  int64(in.URLs),
		Users: int64(in.Users),
	}
}
