// Package compress contains middleware for compressing data transmitted over http.
package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// (1) writer for compress with gzip
// checking http.ResponseWriter interface on gzipWriter
var _ http.ResponseWriter = (*gzipWriter)(nil)

// gzipWriter is the special structure for use in the middleware.
// It implements the ResponseWriter interface.
type gzipWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// NewGzipWriter is the gzipWriter constructor.
func NewGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// three methods of the ResponseWriter interface

// Header implements the ResponseWriter interface method.
func (g *gzipWriter) Header() http.Header {
	return g.w.Header()
}

// Write implements the ResponseWriter interface method.
func (g *gzipWriter) Write(p []byte) (int, error) {
	return g.zw.Write(p)
}

// WriteHeader implements the ResponseWriter interface method.
func (g *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		g.w.Header().Set("Content-Encoding", "gzip")
	}
	g.w.WriteHeader(statusCode)
}

// Close closes gzip.Writer and sends all data from the buffer
func (g *gzipWriter) Close() error {
	return g.zw.Close()
}

// (2) reader for gzip data
// checking io.ReadCloser interface on gzipReader
var _ io.ReadCloser = (*gzipReader)(nil)

// gzipReader is the special structure for use in the middleware.
// It implements the io.ReadCloser interface.
type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewGzipReader is the gzipReader constructor.
func NewGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:  r,
		zr: zr,
	}, nil
}

// two methods if the io.ReadCloser interface

// Read implements the io.ReadCloser interface method.
func (g *gzipReader) Read(p []byte) (n int, err error) {
	return g.zr.Read(p)
}

// Close implements the io.ReadCloser interface method.
func (g *gzipReader) Close() error {
	if err := g.r.Close(); err != nil {
		return err
	}
	return g.zr.Close()
}

// (3) gzip middleware

// gzipContentTypes sets the list of content Types for which the middleware will be used.
var gzipContentTypes = []string{"application/json", "text/html"}

// GzipMiddleware implements a middleware for compressing request body data using gzip.
func GzipMiddleware(h http.Handler) http.Handler {
	gz := func(w http.ResponseWriter, r *http.Request) {
		// firstly set origin http.ResponseWriter
		ow := w // the original http.ResponseWriter

		// checking Content-Type
		contentType := r.Header.Get("Content-Type")
		//supportsContentType := slices.Contains(gzipContentTypes, contentType)
		supportsContentType := contains(gzipContentTypes, contentType)

		// checking that the client is able to receive compressed data in gzip format from the server
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsContentType && supportsGzip {
			gw := NewGzipWriter(w)
			ow = gw
			defer gw.Close()
		}

		// checking that the client has sent compressed data in gzip format to the server
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			gr, err := NewGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = gr // changing the request body to a new one
			defer gr.Close()
		}

		// transferring control to the handler
		h.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(gz)
}

// contains checks for the presence of a row in the list.
func contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
