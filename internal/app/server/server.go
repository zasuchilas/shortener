package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/storage"
	"io"
	"log"
	"net/http"
	"strings"
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
	log.Printf("Server starts at %s", config.ServerAddress)
	log.Fatal(http.ListenAndServe(config.ServerAddress, s.Router()))
}

func (s *Server) Router() chi.Router {
	r := chi.NewRouter()

	// middlewares
	r.Use(middleware.Logger)

	// routes
	r.Post("/", s.writeURLHandler)
	r.Get("/{shortURL}", s.readURLHandler)

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

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	//shortURL = fmt.Sprintf("http://%s/%s", s.outAddr, shortURL)
	shortURL = fmt.Sprintf("%s/%s", config.BaseURL, shortURL)
	if !strings.HasPrefix(shortURL, "http") {
		shortURL = "http://" + shortURL
	}
	_, err = w.Write([]byte(shortURL))
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
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusTemporaryRedirect)
}