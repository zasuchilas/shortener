package shortener

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zasuchilas/shortener/internal/app/model"
	"github.com/zasuchilas/shortener/internal/app/repository"
)

func (s *service) DeleteURLs(ctx context.Context, rawShortURLs []string, userID int64) error {

	// clearing data
	var shortURLs []string
	for _, rawShortURL := range rawShortURLs {
		clean := strings.TrimSpace(rawShortURL)
		if len(clean) == 0 {
			continue
		}
		shortURLs = append(shortURLs, clean)
	}

	// checking request data (1)
	if len(shortURLs) == 0 {
		return fmt.Errorf("the list of short links to delete is empty %w", model.ErrBadRequest)
	}

	// checking request data (2)
	urlCount := len(shortURLs)
	if urlCount > DeletingMaxRowsRequest {
		return fmt.Errorf("the list of short links to delete is too large (actual: %d, maximum: %d) %w", urlCount, DeletingMaxRowsRequest, model.ErrBadRequest)
	}

	// checking request data (3)
	err := s.shortenerRepo.CheckDeletedURLs(ctx, userID, shortURLs)
	if err != nil {
		if errors.Is(err, repository.ErrBadRequest) {
			return fmt.Errorf("%w", model.ErrBadRequest)
		}
		return err
	}

	s.deleteCh <- model.DeleteTask{
		Time:      time.Now(),
		UserID:    userID,
		ShortURLs: shortURLs,
	}

	return nil
}
