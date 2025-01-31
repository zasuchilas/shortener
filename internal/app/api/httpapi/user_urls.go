package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/converter"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/model"
)

// UserURLsHandler is the handler for GET /api/user/urls.
func (i *Implementation) UserURLsHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := GetUserID(r)
	if err != nil {
		logger.Log.Debug("getting userID from ctx", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	out, err := i.shortenerService.UserURLs(r.Context(), userID)
	if err != nil {
		if errors.Is(err, model.ErrNoContent) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	resp := converter.ToHTTPFromUserURL(out)
	if err = enc.Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Log.Debug("sending HTTP 200 response")
}
