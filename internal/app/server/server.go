package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/zasuchilas/shortener/internal/app/storage"
	"io"
	"log"
	"net/http"
)

type Server struct {
	runAddr string
	outAddr string
	db      storage.Storage
}

func New(runAddr, outAddr string, db storage.Storage) *Server {
	return &Server{
		runAddr: runAddr,
		outAddr: outAddr,
		db:      db,
	}
}

func (s *Server) Start() {
	log.Printf("Server starts at %s", s.runAddr)
	log.Fatal(http.ListenAndServe(s.runAddr, s.Router()))
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
	if r.URL.Path != "/" {
		http.Error(w, "unexpected request path", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, "unexpected request method", http.StatusBadRequest)
		return
	}

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
	shortURL = fmt.Sprintf("%s/%s", s.outAddr, shortURL)
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
