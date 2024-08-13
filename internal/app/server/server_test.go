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

//func TestServer_writeURLHandler(t *testing.T) {
//	type want struct {
//		statusCode  int
//		contentType string
//		response    string
//	}
//
//	tests := []struct {
//		name   string
//		method string
//		target string
//		body   string
//		want   want
//	}{
//		{
//			name:   "positive test #1",
//			method: http.MethodPost,
//			target: "/",
//			body:   "http://спорт.ru/",
//			want: want{
//				statusCode:  201,
//				contentType: "text/plain",
//				response:    "http://localhost:8080/", // http://localhost:8080/LcPCiANk
//			},
//		},
//		{
//			name:   "negative test #1 (has many different errors)",
//			method: http.MethodGet,
//			target: "/abc",
//			body:   "http://спор т.ru/",
//			want: want{
//				statusCode:  400,
//				contentType: "text/plain",
//				response:    "", // parse "http://спор т.ru/": invalid character " " in host name
//			},
//		},
//	}
//
//	st := storage.New()
//	srv := New("localhost:8080", st)
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			r := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.body))
//			w := httptest.NewRecorder()
//			srv.writeURLHandler(w, r)
//			res := w.Result()
//
//			// checking status code
//			assert.Equal(t, tt.want.statusCode, res.StatusCode)
//
//			// checking content type
//			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)
//
//			// checking body content
//			defer res.Body.Close()
//			resBody, err := io.ReadAll(res.Body)
//			require.NoError(t, err)
//			assert.Contains(t, string(resBody), tt.want.response)
//		})
//	}
//}

//func TestServer_readURLHandler(t *testing.T) {
//	type want struct {
//		statusCode  int
//		contentType string
//		location    string
//	}
//
//	tests := []struct {
//		name   string
//		method string
//		target string
//		want   want
//	}{
//		{
//			name:   "positive test #1",
//			method: http.MethodGet,
//			target: "http://localhost:8080/abcdefgh",
//			want: want{
//				statusCode:  307,
//				contentType: "text/plain",
//				location:    "http://спорт.ru/",
//			},
//		},
//		{
//			name:   "negative test #1 (has many different errors)",
//			method: http.MethodGet,
//			target: "/abc",
//			want: want{
//				statusCode:  400,
//				contentType: "text/plain",
//				location:    "",
//			},
//		},
//	}
//
//	st := &storage.Database{
//		Urls: make(map[string]string),
//		Hash: make(map[string]string),
//	}
//	srv := New("localhost:8080", st)
//
//	st.Urls["http://спорт.ru/"] = "abcdefgh"
//	st.Hash["abcdefgh"] = "http://спорт.ru/"
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			r := httptest.NewRequest(tt.method, tt.target, nil)
//			w := httptest.NewRecorder()
//			srv.readURLHandler(w, r)
//			res := w.Result()
//			defer res.Body.Close()
//
//			// checking status code
//			assert.Equal(t, tt.want.statusCode, res.StatusCode)
//
//			// checking content type
//			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)
//
//			// checking location header
//
//			assert.Equal(t, res.Header.Get("Location"), tt.want.location)
//			//if tt.want.location != "" {
//			//	assert.NotEmpty(t, res.Header.Get("Location"))
//			//} else {
//			//	assert.Empty(t, res.Header.Get("Location"))
//			//}
//		})
//	}
//}

func TestRouter(t *testing.T) {
	st := &storage.Database{
		Urls: make(map[string]string),
		Hash: make(map[string]string),
	}
	st.Urls["http://спорт.ru/"] = "abcdefgh"
	st.Hash["abcdefgh"] = "http://спорт.ru/"
	srv := New("localhost:8080", st)

	ts := httptest.NewServer(srv.Router())
	defer ts.Close()

	// (1) testing writeURLHandler
	type want1 struct {
		statusCode  int
		contentType string
		response    string
	}

	tests1 := []struct {
		name   string
		method string
		target string
		body   string
		want   want1
	}{
		{
			name:   "positive test #1",
			method: http.MethodPost,
			target: "/",
			body:   "http://спорт.ru/",
			want: want1{
				statusCode:  201,
				contentType: "text/plain",
				response:    "abcdefgh", // "http://localhost:8080/", // http://localhost:8080/LcPCiANk
			},
		},
		{
			name:   "negative test #1 (has many different errors)",
			method: http.MethodGet,
			target: "/abc",
			body:   "http://спор т.ru/",
			want: want1{
				statusCode:  400,
				contentType: "text/plain",
				response:    "", // parse "http://спор т.ru/": invalid character " " in host name
			},
		},
	}

	for _, tt := range tests1 {

		t.Run(tt.name, func(t *testing.T) {
			res, resBody := testRequest(t, ts, tt.method, tt.target, strings.NewReader(tt.body))

			// checking status code
			assert.Equal(t, tt.want.statusCode, res.StatusCode)

			// checking content type
			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)

			// checking body content
			assert.Contains(t, resBody, tt.want.response)
		})
	}

	// (2) testing readURLHandler
	type want2 struct {
		statusCode  int
		contentType string
		location    string
	}

	tests := []struct {
		name   string
		method string
		target string
		want   want2
	}{
		{
			name:   "positive test #1",
			method: http.MethodGet,
			target: "/abcdefgh",
			want: want2{
				statusCode:  307,
				contentType: "text/plain",
				location:    "http://спорт.ru/",
			},
		},
		{
			name:   "negative test #1 (has many different errors)",
			method: http.MethodGet,
			target: "/abc",
			want: want2{
				statusCode:  400,
				contentType: "text/plain",
				location:    "",
			},
		},
	}

	st.Urls["http://спорт.ru/"] = "abcdefgh"
	st.Hash["abcdefgh"] = "http://спорт.ru/"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, _ := testRequest(t, ts, tt.method, tt.target, nil)

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

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Logf("ERR %s", err.Error())
	}
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
