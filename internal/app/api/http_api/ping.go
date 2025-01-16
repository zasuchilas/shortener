package http_api

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/logger"
)

// PingHandler is the handler for GET /ping.
func (i *Implementation) PingHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := i.shortenerService.Ping(ctx); err != nil {
		logger.Log.Debug("postgresql is unavailable", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
