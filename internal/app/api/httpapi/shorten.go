package httpapi

import (
	"encoding/json"
	"github.com/zasuchilas/shortener/internal/app/converter"
	"net/http"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/pkg/shortenerhttpv1"
)

// ShortenHandler is the handler for POST /api/shorten.
func (i *Implementation) ShortenHandler(w http.ResponseWriter, r *http.Request) {

	// getting userID from context
	userID, err := GetUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// decoding request
	var req shortenerhttpv1.ShortenRequest
	dec := json.NewDecoder(r.Body)
	if er := dec.Decode(&req); er != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(er))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	readyURL, conflict, err := i.shortenerService.WriteURL(r.Context(), req.URL, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if conflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	enc := json.NewEncoder(w)
	resp := converter.ToHTTPShortenFromURL(readyURL)
	if err = enc.Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
