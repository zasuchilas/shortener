package shortener

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/model"
	"github.com/zasuchilas/shortener/internal/app/utils/urlfuncs"
)

func (s *service) ShortenBatch(ctx context.Context, in []model.ShortenBatchIn, userID int64) (out []model.ShortenBatchOut, err error) {

	// checking request data
	wrongBatchItems := make([]string, 0)
	for i, item := range in {
		origURL, e := urlfuncs.CleanURL(item.OriginalURL)
		if e != nil {
			rowErr := fmt.Sprintf("Pos: %d, correlation_id: \"%s\", original_url: \"%s\", error: \"%s\"",
				i, item.CorrelationID, item.OriginalURL, e.Error())
			wrongBatchItems = append(wrongBatchItems, rowErr)
			continue
		}
		in[i].OriginalURL = origURL
	}
	if len(wrongBatchItems) > 0 {
		return nil, fmt.Errorf("wrong batch items: %s", strings.Join(wrongBatchItems, ", "))
	}

	// getting origURLs for query
	origURLs := make([]string, 0)
	for _, item := range in {
		origURLs = append(origURLs, item.OriginalURL)
	}

	start := time.Now()
	logger.Log.Info("batching data starting", zap.Time("start", start))

	urlRows, err := s.shortenerRepo.WriteURLs(ctx, origURLs, userID)
	if err != nil {
		return nil, fmt.Errorf("batching urls %w", err)
	}

	end := time.Now()
	logger.Log.Info("batching data ending",
		zap.Duration("duration", time.Since(start)),
		zap.Time("end", end))

	out = make([]model.ShortenBatchOut, len(urlRows))
	for i, requestItem := range in {
		out[i] = model.ShortenBatchOut{
			CorrelationID: requestItem.CorrelationID,
			ShortURL:      urlfuncs.EnrichURL(urlRows[requestItem.OriginalURL].ShortURL),
		}
	}

	return out, nil
}
