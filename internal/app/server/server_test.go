package server

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zasuchilas/shortener/internal/app/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_writeURLHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		response    string
	}

	tests := []struct {
		name   string
		method string
		target string
		body   string
		want   want
	}{
		{
			name:   "positive test #1",
			method: http.MethodPost,
			target: "/",
			body:   "http://спорт.ru/",
			want: want{
				statusCode:  201,
				contentType: "text/plain",
				response:    "http://localhost:8080/", // http://localhost:8080/LcPCiANk
			},
		},
		{
			name:   "negative test #1 (has many different errors)",
			method: http.MethodGet,
			target: "/abc",
			body:   "http://спор т.ru/",
			want: want{
				statusCode:  400,
				contentType: "text/plain",
				response:    "", // parse "http://спор т.ru/": invalid character " " in host name
			},
		},
	}

	st := storage.New()
	srv := New("localhost:8080", st)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			srv.writeURLHandler(w, r)
			res := w.Result()

			// checking status code
			assert.Equal(t, tt.want.statusCode, res.StatusCode)

			// checking content type
			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)

			// checking body content
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Contains(t, string(resBody), tt.want.response)
		})
	}
}

func TestServer_readURLHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		location    string
	}

	tests := []struct {
		name   string
		method string
		target string
		want   want
	}{
		{
			name:   "positive test #1",
			method: http.MethodGet,
			target: "http://localhost:8080/abcdefgh",
			want: want{
				statusCode:  307,
				contentType: "text/plain",
				location:    "http://спорт.ru/",
			},
		},
		{
			name:   "negative test #1 (has many different errors)",
			method: http.MethodGet,
			target: "/abc",
			want: want{
				statusCode:  400,
				contentType: "text/plain",
				location:    "",
			},
		},
	}

	st := &storage.Database{
		Urls: make(map[string]string),
		Hash: make(map[string]string),
	}
	srv := New("localhost:8080", st)

	st.Urls["http://спорт.ru/"] = "abcdefgh"
	st.Hash["abcdefgh"] = "http://спорт.ru/"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.target, nil)
			w := httptest.NewRecorder()
			srv.readURLHandler(w, r)
			res := w.Result()

			// checking status code
			assert.Equal(t, tt.want.statusCode, res.StatusCode)

			// checking content type
			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)

			// checking location header

			assert.Equal(t, res.Header.Get("Location"), tt.want.location)
			//if tt.want.location != "" {
			//	assert.NotEmpty(t, res.Header.Get("Location"))
			//} else {
			//	assert.Empty(t, res.Header.Get("Location"))
			//}
		})
	}
}
