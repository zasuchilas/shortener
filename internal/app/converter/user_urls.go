package converter

import (
	"github.com/zasuchilas/shortener/internal/app/model"
	"github.com/zasuchilas/shortener/pkg/shortener_http_api_v1"
)

func ToHTTPFromUserURL(in []model.UserURL) []shortener_http_api_v1.UserURLsResponseItem {
	result := make([]shortener_http_api_v1.UserURLsResponseItem, len(in))
	for i := range in {
		result[i] = shortener_http_api_v1.UserURLsResponseItem{
			ShortURL:    in[i].ShortURL,
			OriginalURL: in[i].OriginalURL,
		}
	}
	return result
}
