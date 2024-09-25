package middleware

import (
	"github.com/zasuchilas/shortener/internal/app/compress"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"net/http"
	"strings"
)

var gzipContentTypes = []string{"application/json", "text/html"}

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
