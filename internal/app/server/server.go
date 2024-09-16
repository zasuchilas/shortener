package server

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"github.com/zasuchilas/shortener/internal/app/storage"
	"github.com/zasuchilas/shortener/internal/app/storage/urlfuncs"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type Server struct {
	store storage.Storage
}

func New(s storage.Storage) *Server {
	return &Server{
		store: s,
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
	r.Use(GzipMiddleware)

	// routes
	r.Post("/", s.writeURLHandler)
	r.Get("/{shortURL}", s.readURLHandler)
	r.Post("/api/shorten", s.shortenHandler)
	r.Post("/api/shorten/batch", s.shortenBatchHandler)
	r.Get("/ping", s.ping)

	return r
}

func (s *Server) Stop() {}

func (s *Server) writeURLHandler(w http.ResponseWriter, r *http.Request) {

	logger.Log.Debug("decoding request")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Debug("checking request data")
	rawURL := string(body)
	origURL, err := urlfuncs.CleanURL(rawURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Debug("performing the endpoint task")
	shortURL, _, err := s.store.WriteURL(r.Context(), origURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := urlfuncs.EnrichURL(shortURL)

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

	origURL, err := s.store.ReadURL(r.Context(), shortURL)
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

	logger.Log.Debug("checking request data")
	origURL, err := urlfuncs.CleanURL(req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Debug("performing the endpoint task")
	shortURL, _, err := s.store.WriteURL(r.Context(), origURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Debug("filling in the response model")
	resp := models.ShortenResponse{
		Result: urlfuncs.EnrichURL(shortURL),
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

func (s *Server) shortenBatchHandler(w http.ResponseWriter, r *http.Request) {

	logger.Log.Debug("decoding request")
	var req models.ShortenBatchRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	start := time.Now()
	logger.Log.Debug("batching data starting", zap.Time("start", start))

	// transaction + stored + named
	// tx err -> err bundle 5 tries
	// writeURL change to multi params (maps/file already use tx <- change retry placing)
	// indexes to db

	// pre
	// create dataset for work
	// invalid items (fill empty)
	// double items (skip)
	// already used items (fill)

	end := time.Now()
	logger.Log.Debug("batching data ending",
		zap.Duration("duration", time.Since(start)),
		zap.Time("end", end))

	resp := req

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	logger.Log.Debug("encoding response")
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Log.Debug("sending HTTP 201 response")
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := s.store.Ping(ctx); err != nil {
		logger.Log.Debug("postgresql is unavailable", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
