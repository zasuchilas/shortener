package shortener

import (
	"context"
	"github.com/zasuchilas/shortener/internal/app/utils/urlfuncs"
)

func (s *service) WriteURL(ctx context.Context, rawURL string, userID int64) (readyURL string, conflict bool, err error) {

	// checking request data
	origURL, err := urlfuncs.CleanURL(rawURL)
	if err != nil {
		// TODO: bad request
		return "", false, err
	}

	// performing the endpoint task
	shortURL, conflict, err := s.shortenerRepo.WriteURL(ctx, origURL, userID)
	if err != nil {
		return "", false, err
	}
	readyURL = urlfuncs.EnrichURL(shortURL)

	return readyURL, conflict, err
}
