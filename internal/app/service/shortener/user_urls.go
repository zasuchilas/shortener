package shortener

import (
	"context"
	"errors"
	"fmt"

	"github.com/zasuchilas/shortener/internal/app/model"
	"github.com/zasuchilas/shortener/internal/app/repository"
	"github.com/zasuchilas/shortener/internal/app/utils/urlfuncs"
)

func (s *service) UserURLs(ctx context.Context, userID int64) (out []model.UserURL, err error) {

	urlRowList, err := s.shortenerRepo.UserURLs(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w", model.ErrNoContent)
		}
		return nil, err
	}

	out = make([]model.UserURL, len(urlRowList))
	for i, row := range urlRowList {
		out[i] = model.UserURL{
			ShortURL:    urlfuncs.EnrichURL(row.ShortURL),
			OriginalURL: row.OrigURL,
		}
	}

	return out, nil
}
