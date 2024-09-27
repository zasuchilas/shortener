package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// (1) writer for compress with gzip

type gzipWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func NewGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// three methods of the ResponseWriter interface

func (g *gzipWriter) Header() http.Header {
	return g.w.Header()
}

func (g *gzipWriter) Write(p []byte) (int, error) {
	return g.zw.Write(p)
}

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

type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

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

func (g *gzipReader) Read(p []byte) (n int, err error) {
	return g.zr.Read(p)
}

func (g *gzipReader) Close() error {
	if err := g.r.Close(); err != nil {
		return err
	}
	return g.zr.Close()
}

// (3) gzip middleware

var gzipContentTypes = []string{"application/json", "text/html"}

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

func contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
