package compress

import (
	"compress/gzip"
	"io"
	"net/http"
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
