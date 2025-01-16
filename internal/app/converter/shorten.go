package converter

import (
	"github.com/zasuchilas/shortener/pkg/shortener_http_api_v1"
)

func ToHTTPShortenFromURL(readyURL string) shortener_http_api_v1.ShortenResponse {
	return shortener_http_api_v1.ShortenResponse{
		Result: readyURL,
	}
}
