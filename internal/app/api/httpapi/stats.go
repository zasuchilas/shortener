package httpapi

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/converter"
	"github.com/zasuchilas/shortener/internal/app/logger"
)

// StatsHandler is the handler for GET /api/internal/stats.
func (i *Implementation) StatsHandler(w http.ResponseWriter, r *http.Request) {

	out, err := i.shortenerService.Stats(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	resp := converter.ToHTTPFromStats(*out)
	if err = enc.Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
