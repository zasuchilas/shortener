package httpserver

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/secure"
	"github.com/zasuchilas/shortener/internal/app/utils/compress"
	"github.com/zasuchilas/shortener/pkg/shortenerhttpv1"
	"github.com/zasuchilas/shortener/pkg/trusted"
)

// Server _
type Server struct {
	server  http.Server
	secure  *secure.Secure
	httpAPI shortenerhttpv1.ShortenerHTTPApiV1
}

// NewServer _
func NewServer(httpAPI shortenerhttpv1.ShortenerHTTPApiV1, secure *secure.Secure) *Server {
	return &Server{
		httpAPI: httpAPI,
		secure:  secure,
	}
}

// Run starts http Server.
func (s *Server) Run() {
	logger.Log.Info("Server starts", zap.String("addr", config.ServerAddress))
	s.server = http.Server{
		Addr:    config.ServerAddress,
		Handler: s.Router(),
	}
	var err error

	if !config.EnableHTTPS {
		// running http Server
		if err = s.server.ListenAndServe(); err != http.ErrServerClosed {
			// listener start or stop errors
			logger.Log.Panic("HTTP Server ListenAndServe", zap.String("err", err.Error()))
		}
	} else {
		// creating cert.pem and key.pem
		err = makePemFiles()
		if err != nil {
			logger.Log.Panic("making pem files for TLS", zap.String("err", err.Error()))
		}
		// running https Server
		if err = s.server.ListenAndServeTLS("./cert.pem", "./key.pem"); err != http.ErrServerClosed {
			// listener start or stop errors
			logger.Log.Panic("HTTPS Server ListenAndServeTLS", zap.String("err", err.Error()))
		}
	}
}

// Stop stops http Server.
func (s *Server) Stop() {
	// stopping the Server
	if err := s.server.Shutdown(context.Background()); err != nil {
		// listener closing errors
		log.Printf("HTTP Server Shutdown: %v", err)
	}
}

// Router sets the routes.
//
// A chi router is used: https://github.com/go-chi/chi/
func (s *Server) Router() chi.Router {
	r := chi.NewRouter()

	// middlewares
	r.Use(middleware.Logger)
	//r.Use(logger.LoggingMiddleware)
	r.Use(compress.GzipMiddleware)
	r.Mount("/debug/", middleware.Profiler())

	// routes
	r.Get("/{shortURL}", s.httpAPI.ReadURLHandler)
	r.Get("/ping", s.httpAPI.PingHandler)

	// routes with guard (if there is no valid token returns error 401 Unauthorized)
	r.Group(func(r chi.Router) {
		r.Use(s.secure.GuardMiddleware)
		r.Get("/api/user/urls", s.httpAPI.UserURLsHandler)
		r.Delete("/api/user/urls", s.httpAPI.DeleteURLsHandler)
	})

	// routes with secure cookie (if there is no valid token assigns a new token)
	r.Group(func(r chi.Router) {
		r.Use(s.secure.SecureMiddleware)
		r.Post("/", s.httpAPI.WriteURLHandler)
		r.Post("/api/shorten", s.httpAPI.ShortenHandler)
		r.Post("/api/shorten/batch", s.httpAPI.ShortenBatchHandler)
	})

	trustedSubnet := trusted.NewTrustedSubnet(config.TrustedSubnet)
	r.Group(func(r chi.Router) {
		r.Use(trustedSubnet.Middleware)
		r.Get("/api/internal/stats", s.httpAPI.StatsHandler)
	})

	return r
}
