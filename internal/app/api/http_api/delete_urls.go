package http_api

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/model"
)

// DeleteURLsHandler is the handler for DELETE /api/user/urls.
func (i *Implementation) DeleteURLsHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := GetUserID(r)
	if err != nil {
		logger.Log.Debug("getting userID from ctx", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// decoding request
	var rawShortURLs []string
	dec := json.NewDecoder(r.Body)
	if err = dec.Decode(&rawShortURLs); err != nil {
		logger.Log.Info("cannot decode request JSON body", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = i.shortenerService.DeleteURLs(r.Context(), rawShortURLs, userID)
	if err != nil {
		if errors.Is(err, model.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
