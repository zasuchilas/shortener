package httpapi

import (
	"io"
	"net/http"
)

// WriteURLHandler is the handler for POST /.
func (i *Implementation) WriteURLHandler(w http.ResponseWriter, r *http.Request) {

	// getting userID from context
	userID, err := GetUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// decoding request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	readyURL, conflict, err := i.shortenerService.WriteURL(r.Context(), string(body), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if conflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	_, err = w.Write([]byte(readyURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
