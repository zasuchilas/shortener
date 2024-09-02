package server

import (
	"github.com/zasuchilas/shortener/internal/app/compress"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

var gzipContentTypes = []string{"application/json", "text/html"}

type (
	// structure for storing information about the response
	responseData struct {
		status int
		size   int
	}

	// http.ResponseWriter implementation
	loggingResponseWriter struct {
		http.ResponseWriter // use original http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// we record the response using the original http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // take the size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// we record the status code using the original http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // take the status code
}

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)
		logger.Log.Info(
			"HTTP REQUEST",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Duration("duration", duration),
			zap.Int("size", responseData.size),
		)
	}
	return http.HandlerFunc(logFn)
}

func GzipMiddleware(h http.Handler) http.Handler {
	gz := func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("firstly set origin http.ResponseWriter")
		ow := w // the original http.ResponseWriter

		logger.Log.Debug("checking Content-Type")
		contentType := r.Header.Get("Content-Type")
		//supportsContentType := slices.Contains(gzipContentTypes, contentType)
		supportsContentType := contains(gzipContentTypes, contentType)

		logger.Log.Debug("checking that the client is able to receive compressed data in gzip format from the server")
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsContentType && supportsGzip {
			gw := compress.NewGzipWriter(w)
			ow = gw
			defer gw.Close()
		}

		logger.Log.Debug("checking that the client has sent compressed data in gzip format to the server")
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			gr, err := compress.NewGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = gr // changing the request body to a new one
			defer gr.Close()
		}

		logger.Log.Debug("transferring control to the handler")
		h.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(gz)
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
