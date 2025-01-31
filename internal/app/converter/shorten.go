package converter

import (
	"github.com/zasuchilas/shortener/pkg/shortenerhttpv1"
)

func ToHTTPShortenFromURL(readyURL string) shortenerhttpv1.ShortenResponse {
	return shortenerhttpv1.ShortenResponse{
		Result: readyURL,
	}
}
