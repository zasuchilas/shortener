package server

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zasuchilas/shortener/internal/app/secure"
	"github.com/zasuchilas/shortener/internal/app/storage"
)

var (
	st  *storage.DBMaps
	sec *secure.Secure
	s   *Server
	srv *httptest.Server
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	log.Println("setup started")

	st = storage.NewDBMaps()
	//refillDatabase()

	sec = secure.New("supersecretkey", "", "")
	s = New(st, sec)
	srv = httptest.NewServer(s.Router())

	log.Println("setup completed")
}

func teardown() {
	srv.Close()
	log.Println("teardown completed")
}

//func refillDatabase() {
//	storage.Write(st, 1, 1, "abcdefgh", "http://спорт.ru/")
//}

func testRequest(t *testing.T, method,
	path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, srv.URL+path, body)
	require.NoError(t, err)

	client := srv.Client()
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
				statusCode:  201, // 409
				contentType: "text/plain",
				response:    "19xtf1ts", // "http://localhost:8080/", // http://localhost:8080/LcPCiANk
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

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			res, resBody := testRequest(t, tt.method, tt.target, strings.NewReader(tt.body))
			defer res.Body.Close()

			// checking status code
			assert.Equal(t, tt.want.statusCode, res.StatusCode)

			// checking content type
			assert.Contains(t, res.Header.Get("Content-Type"), tt.want.contentType)

			// checking body content
			assert.Contains(t, resBody, tt.want.response)
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
			target: "/19xtf1ts",
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

	//refillDatabase()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, _ := testRequest(t, tt.method, tt.target, nil)
			defer res.Body.Close()

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

func TestServer_shortenHandler(t *testing.T) {
	const url = "/api/shorten"

	tests := []struct {
		name                string
		method              string
		url                 string
		body                string
		expectedCode        int
		expectedBody        string
		expectedContentType string
	}{
		{
			name:         "method_get",
			method:       http.MethodGet,
			url:          url,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "method_put",
			method:       http.MethodPut,
			url:          url,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "method_post_without_body",
			method:       http.MethodPost,
			url:          url,
			body:         "",
			expectedCode: http.StatusInternalServerError,
			expectedBody: "",
		},
		//{
		//	name:                "positive #1",
		//	method:              http.MethodPost,
		//	url:                 url,
		//	body:                `{"url": "http://спорт.ru/"}`,
		//	expectedCode:        http.StatusConflict, // http.StatusCreated,
		//	expectedBody:        fmt.Sprintf(`{"result": "http://%s/abcdefgh"}`, config.BaseURL),
		//	expectedContentType: "application/json",
		//},
	}

	for _, tc := range tests {
		// create request with resty
		req := resty.New().R()
		req.Method = tc.method
		req.URL = srv.URL + tc.url
		if len(tc.body) > 0 {
			req.SetHeader("Content-Type", "application/json")
			req.SetBody(tc.body)
		}

		// execute request
		resp, err := req.Send()
		assert.NoError(t, err, "error making HTTP request")

		// checking status code
		assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Response code didn't match expected")

		// checking body
		if tc.expectedBody != "" {
			assert.JSONEq(t, tc.expectedBody, string(resp.Body()))
		}

		// checking content-type header
		assert.Contains(t, resp.Header().Get("Content-Type"), tc.expectedContentType)

	}
}

func TestGzipCompression(t *testing.T) {
	// TODO: TestGzipCompression
	//handler := http.HandlerFunc(GzipMiddleware(s.shortenHandler))
}
