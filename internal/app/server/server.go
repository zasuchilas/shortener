package server

import (
	"fmt"
	"github.com/zasuchilas/shortener/internal/app/storage"
	"io"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	addr string
	db   storage.Storage
}

func New(addr string, db storage.Storage) *Server {
	return &Server{
		addr: addr,
		db:   db,
	}
}

func (s *Server) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.writeURLHandler)
	mux.HandleFunc("/{shortURL}", s.readURLHandler)

	log.Printf("Server starts at %s", s.addr)
	err := http.ListenAndServe(s.addr, mux)
	if err != nil {
		panic(any(err))
	}
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

	shortURL = fmt.Sprintf("http://%s/%s", s.addr, shortURL)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) readURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO: shortURL := r.PathValue("shortURL") (NOT WORK IN httptest)
	shortURL := r.URL.Path
	shortURL = strings.TrimLeft(shortURL, "/")
	log.Println("shortURL 1", shortURL)

	origURL, err := s.db.ReadURL(shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", origURL)
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
