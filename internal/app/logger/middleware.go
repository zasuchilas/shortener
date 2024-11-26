package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	// responseData is the structure for storing information about the response.
	responseData struct {
		status int
		size   int
	}

	// loggingResponseWriter is the http.ResponseWriter implementation.
	loggingResponseWriter struct {
		http.ResponseWriter // use original http.ResponseWriter
		responseData        *responseData
	}
)

// Write implements method Write for http.ResponseWriter interface.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// we record the response using the original http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // take the size
	return size, err
}

// WriteHeader implements method WriteHeader for http.ResponseWriter interface.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// we record the status code using the original http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // take the status code
}

// LoggingMiddleware implements the logging middleware.
func LoggingMiddleware(h http.Handler) http.Handler {
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
		Log.Info(
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
