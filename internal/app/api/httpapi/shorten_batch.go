package httpapi

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/converter"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/pkg/shortenerhttpv1"
)

// ShortenBatchHandler is the handler for POST /api/shorten/batch.
func (i *Implementation) ShortenBatchHandler(w http.ResponseWriter, r *http.Request) {

	// getting userID from context
	userID, err := GetUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// decoding request
	var req []shortenerhttpv1.ShortenBatchRequestItem
	dec := json.NewDecoder(r.Body)
	if err = dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	out, err := i.shortenerService.ShortenBatch(r.Context(), converter.ToShortenBatchInFromHTTP(req), userID)
	if err != nil {
		logger.Log.Debug("shorten batch", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	// encoding response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(w)
	resp := converter.ToHTTPFromShortenBatchOut(out)
	if err = enc.Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
