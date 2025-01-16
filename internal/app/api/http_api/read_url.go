package http_api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zasuchilas/shortener/internal/app/repository"
)

// ReadURLHandler is the handler for GET /{shortURL}.
func (i *Implementation) ReadURLHandler(w http.ResponseWriter, r *http.Request) {

	shortURL := chi.URLParam(r, "shortURL")

	origURL, err := i.shortenerService.ReadURL(r.Context(), shortURL)
	if err != nil {
		if errors.Is(err, repository.ErrGone) {
			http.Error(w, err.Error(), http.StatusGone)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest) // according to the assignment, so, but postgresql may give an internal error
		return
	}

	w.Header().Set("Location", origURL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
