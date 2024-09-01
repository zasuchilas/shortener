package server

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"github.com/zasuchilas/shortener/internal/app/storage"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type Server struct {
	db storage.Storage
}

func New(db storage.Storage) *Server {
	return &Server{
		db: db,
	}
}

func (s *Server) Start() {
	logger.Log.Info("Server starts", zap.String("addr", config.ServerAddress))
	logger.Log.Fatal(http.ListenAndServe(config.ServerAddress, s.Router()).Error())
}

func (s *Server) Router() chi.Router {
	r := chi.NewRouter()

	// middlewares
	r.Use(WithLogging) // r.Use(middleware.Logger)

	// routes
	r.Post("/", s.writeURLHandler)
	r.Get("/{shortURL}", s.readURLHandler)
	r.Post("/api/shorten", s.shortenHandler)

	return r
}

// TODO: func (s *Server) Stop() {}

func (s *Server) writeURLHandler(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rawURL := string(body)
	shortURL, err := s.db.WriteURL(rawURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := storage.EnrichURL(shortURL)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(resp))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) readURLHandler(w http.ResponseWriter, r *http.Request) {

	shortURL := chi.URLParam(r, "shortURL")

	origURL, err := s.db.ReadURL(shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", origURL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) shortenHandler(w http.ResponseWriter, r *http.Request) {

	logger.Log.Debug("decoding request")
	var req models.ShortenRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Log.Debug("performing the endpoint task")
	shortURL, err := s.db.WriteURL(req.Url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Debug("filling in the response model")
	resp := models.ShortenResponse{
		Result: storage.EnrichURL(shortURL),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	logger.Log.Debug("encoding response")
	enc := json.NewEncoder(w)
	if err = enc.Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		// TODO: ? http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Log.Debug("sending HTTP 201 response")
}
