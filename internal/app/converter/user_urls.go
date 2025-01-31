package converter

import (
	"github.com/zasuchilas/shortener/internal/app/model"
	"github.com/zasuchilas/shortener/pkg/shortenerhttpv1"
)

func ToHTTPFromUserURL(in []model.UserURL) []shortenerhttpv1.UserURLsResponseItem {
	result := make([]shortenerhttpv1.UserURLsResponseItem, len(in))
	for i := range in {
		result[i] = shortenerhttpv1.UserURLsResponseItem{
			ShortURL:    in[i].ShortURL,
			OriginalURL: in[i].OriginalURL,
		}
	}
	return result
}
