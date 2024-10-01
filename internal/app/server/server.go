package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"github.com/zasuchilas/shortener/internal/app/secure"
	"github.com/zasuchilas/shortener/internal/app/storage"
	"github.com/zasuchilas/shortener/internal/app/utils/compress"
	"github.com/zasuchilas/shortener/internal/app/utils/urlfuncs"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	secure *secure.Secure

	store    storage.Storage
	deleteCh chan models.DeleteTask
}

func New(s storage.Storage, secure *secure.Secure) *Server {
	srv := &Server{
		secure: secure,
		store:  s,
	}

	// batch deleting
	srv.deleteCh = make(chan models.DeleteTask, storage.DeletingChanBuffer)
	go srv.flushDeletingTasks()

	return srv
}

func (s *Server) Start() {
	logger.Log.Info("Server starts", zap.String("addr", config.ServerAddress))
	logger.Log.Fatal(http.ListenAndServe(config.ServerAddress, s.Router()).Error())
}

func (s *Server) Router() chi.Router {
	r := chi.NewRouter()

	// middlewares
	r.Use(logger.LoggingMiddleware) // r.Use(middleware.Logger)
	r.Use(compress.GzipMiddleware)

	// routes
	r.Get("/{shortURL}", s.readURLHandler)
	r.Get("/ping", s.ping)

	// routes with guard (if there is no valid token returns error 401 Unauthorized)
	r.Group(func(r chi.Router) {
		r.Use(s.secure.GuardMiddleware)
		r.Get("/api/user/urls", s.userURLsHandler)
		r.Delete("/api/user/urls", s.deleteURLsHandler)
	})

	// routes with secure cookie (if there is no valid token assigns a new token)
	r.Group(func(r chi.Router) {
		r.Use(s.secure.SecureMiddleware)
		r.Post("/", s.writeURLHandler)
		r.Post("/api/shorten", s.shortenHandler)
		r.Post("/api/shorten/batch", s.shortenBatchHandler)
	})

	return r
}

func (s *Server) Stop() {}

func (s *Server) writeURLHandler(w http.ResponseWriter, r *http.Request) {

	// getting userID from context
	userID, err := getUserID(r)
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

	// checking request data
	rawURL := string(body)
	origURL, err := urlfuncs.CleanURL(rawURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// performing the endpoint task
	shortURL, conflict, err := s.store.WriteURL(r.Context(), origURL, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := urlfuncs.EnrichURL(shortURL)
	w.Header().Set("Content-Type", "text/plain")
	if conflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

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
		if errors.Is(err, storage.ErrGone) {
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

func (s *Server) shortenHandler(w http.ResponseWriter, r *http.Request) {

	// getting userID from context
	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// decoding request
	var req models.ShortenRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// checking request data
	origURL, err := urlfuncs.CleanURL(req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// performing the endpoint task
	shortURL, conflict, err := s.store.WriteURL(r.Context(), origURL, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Log.Debug("filling in the response model")
	resp := models.ShortenResponse{
		Result: urlfuncs.EnrichURL(shortURL),
	}

	w.Header().Set("Content-Type", "application/json")
	if conflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

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

	// getting userID from context
	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// decoding request
	var req models.ShortenBatchRequest
	dec := json.NewDecoder(r.Body)
	if err = dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// checking request data
	wrongBatchItems := make([]string, 0)
	for i, item := range req {
		origURL, e := urlfuncs.CleanURL(item.OriginalURL)
		if e != nil {
			rowErr := fmt.Sprintf("Pos: %d, correlation_id: \"%s\", original_url: \"%s\", error: \"%s\"",
				i, item.CorrelationID, item.OriginalURL, e.Error())
			wrongBatchItems = append(wrongBatchItems, rowErr)
			continue
		}
		req[i].OriginalURL = origURL
	}
	if len(wrongBatchItems) > 0 {
		http.Error(w, strings.Join(wrongBatchItems, ", "), http.StatusBadRequest)
		return
	}

	logger.Log.Debug("creating origURLs for query", zap.Int("len(request)", len(req)))
	origURLs := make([]string, 0)
	for _, item := range req {
		origURLs = append(origURLs, item.OriginalURL)
	}
	logger.Log.Debug("origURLs created", zap.Int("len(origURLs)", len(origURLs)))
	if len(req) != len(origURLs) {
		logger.Log.Error("len(req) != len(origURLs)")
	}

	start := time.Now()
	logger.Log.Debug("batching data starting", zap.Time("start", start))

	urlRows, err := s.store.WriteURLs(r.Context(), origURLs, userID)
	if err != nil {
		logger.Log.Error("writing URLs error", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	end := time.Now()
	logger.Log.Debug("batching data ending",
		zap.Duration("duration", time.Since(start)),
		zap.Time("end", end))

	logger.Log.Debug("creating success response")
	resp := make(models.ShortenBatchResponse, len(req))
	for i, requestItem := range req {
		resp[i] = models.ShortenBatchResponseItem{
			CorrelationID: requestItem.CorrelationID,
			ShortURL:      urlfuncs.EnrichURL(urlRows[requestItem.OriginalURL].ShortURL),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// encoding response
	enc := json.NewEncoder(w)
	if err = enc.Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.String("error", err.Error()))
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

func (s *Server) deleteURLsHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := getUserID(r)
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

	// clearing data
	var shortURLs []string
	for _, rawShortURL := range rawShortURLs {
		clean := strings.TrimSpace(rawShortURL)
		if len(clean) == 0 {
			continue
		}
		shortURLs = append(shortURLs, clean)
	}

	// checking request data (1)
	if len(shortURLs) == 0 {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("the list of short links to delete is empty"))
		return
	}

	// checking request data (2)
	urlCount := len(shortURLs)
	if urlCount > storage.DeletingMaxRowsRequest {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		msg := fmt.Sprintf("the list of short links to delete is too large (actual: %d, maximum: %d)", urlCount, storage.DeletingMaxRowsRequest)
		_, _ = w.Write([]byte(msg))
		return
	}

	// checking request data (3)
	err = s.store.CheckDeletedURLs(r.Context(), userID, shortURLs)
	if err != nil {
		if errors.Is(err, storage.ErrBadRequest) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.deleteCh <- models.DeleteTask{
		Time:      time.Now(),
		UserID:    userID,
		ShortURLs: shortURLs,
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) userURLsHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := getUserID(r)
	if err != nil {
		logger.Log.Debug("getting userID from ctx", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	urlRowList, err := s.store.UserURLs(r.Context(), userID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := make(models.UserURLsResponse, len(urlRowList))
	for i, row := range urlRowList {
		resp[i] = models.UserURLsResponseItem{
			ShortURL:    urlfuncs.EnrichURL(row.ShortURL),
			OriginalURL: row.OrigURL,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err = enc.Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Log.Debug("sending HTTP 200 response")
}

// getUserID gets the userID from the context.
// All errors in this method are considered internal (500 InternalServerError)
// because error 401 Unauthorized is returned earlier from middleware.
func getUserID(r *http.Request) (userID int64, err error) {

	// getting userID from context of request (after SecureMiddleware)
	uid := r.Context().Value(secure.ContextUserIDKey)

	// cast userID from any to int64
	userID, err = strconv.ParseInt(fmt.Sprintf("%d", uid), 10, 64)
	if err != nil {
		logger.Log.Debug("there are problems with userID", zap.Error(err))
		return 0, err
	}
	if userID == 0 {
		logger.Log.Debug("something went wrong: empty userID")
		return 0, errors.New("something went wrong: empty userID")
	}

	return userID, nil
}

func (s *Server) flushDeletingTasks() {

	// the interval for sending data to the database
	ticker := time.NewTicker(storage.DeletingFlushInterval)

	var shortURLs []string
	// TODO: use generator & buffer chan for limit shortURLs slice
	// channel for closing
	//doneCh := make(chan struct{})
	//defer close(doneCh)

	for {
		select {
		case task := <-s.deleteCh:
			shortURLs = append(shortURLs, task.ShortURLs...)
			//inputCh := deleteGenerator(doneCh, task)
		case <-ticker.C:
			// if there is nothing to send, we do not send anything
			if len(shortURLs) == 0 {
				continue
			}

			err := s.store.DeleteURLs(context.TODO(), shortURLs...)
			if err != nil {
				logger.Log.Info("cannot delete urls",
					zap.String("error", err.Error()), zap.String("shortURLs", strings.Join(shortURLs, ", ")))

				// we will try to delete the data next time
				continue
			}

			// clearing the deletion queue
			shortURLs = nil
		}
	}
}

// TODO: ... learning is good, but there is the KISS
//func deleteGenerator(doneCh chan struct{}, task models.DeleteTask) chan string {
//	inputCh := make(chan string)
//
//	go func() {
//		defer  close(inputCh)
//
//		for _, shortURL := range task.ShortURLs {
//			select {
//			case <-doneCh:
//				return
//			case inputCh <- shortURL:
//			}
//		}
//	}()
//
//	return inputCh
//}
