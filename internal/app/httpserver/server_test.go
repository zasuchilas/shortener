package httpserver

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/zasuchilas/shortener/internal/app/api/httpapi"
	"github.com/zasuchilas/shortener/internal/app/service"
	"github.com/zasuchilas/shortener/internal/app/service/shortener"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/repository"
	"github.com/zasuchilas/shortener/internal/app/secure"
)

var (
	shortenerRepo    *repository.DBMaps
	secureService    *secure.Secure
	shortenerService service.ShortenerService
	httpServer       *Server
	testServer       *httptest.Server
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	shortenerRepo = repository.NewDBMaps()
	secureService = secure.New("supersecretkey", "", "")
	shortenerService = shortener.NewService(shortenerRepo, secureService)
	httpServer = NewServer(httpapi.NewImplementation(shortenerService), secureService)
	testServer = httptest.NewServer(httpServer.Router())
}

func teardown() {
	testServer.Close()
	log.Println("teardown completed")
}

//func refillDatabase() {
//	storage.Write(shortenerRepo, 1, 1, "abcdefgh", "http://спорт.ru/")
//}

func testRequest(t *testing.T, method,
	path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, testServer.URL+path, body)
	require.NoError(t, err)

	client := testServer.Client()
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
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:         "wrong url",
			method:       http.MethodPost,
			url:          url,
			body:         "http://сп  орт.ru/",
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:                "positive #1",
			method:              http.MethodPost,
			url:                 url,
			body:                `{"url": "http://спорт.ru/"}`,
			expectedCode:        http.StatusCreated, // http.StatusConflict,
			expectedBody:        fmt.Sprintf(`{"result": "http://%s/19xtf1ts"}`, config.BaseURL),
			expectedContentType: "application/json",
		},
	}

	for _, tc := range tests {
		setup()

		// create request with resty
		req := resty.New().R()
		req.Method = tc.method
		req.URL = testServer.URL + tc.url
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

		testServer.Close()
	}
}

func TestServer_shortenBatchHandler(t *testing.T) {
	const url = "/api/shorten/batch"

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
			name:                "positive #1",
			method:              http.MethodPost,
			url:                 url,
			body:                `[{"correlation_id": "batch1", "original_url": "https://ya.ru"}]`,
			expectedCode:        http.StatusCreated,
			expectedBody:        fmt.Sprintf(`[{"correlation_id": "batch1",	"short_url": "http://%s/19xtf1ts"}]`, config.BaseURL),
			expectedContentType: "application/json",
		},
	}

	for _, tc := range tests {
		setup()

		// create request with resty
		req := resty.New().R()
		req.Method = tc.method
		req.URL = testServer.URL + tc.url
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

		testServer.Close()
	}
}

func TestServer_deleteURLsHandler(t *testing.T) {
	const url = "/api/user/urls"
	setup()

	// create URL request
	req1 := resty.New().R()
	req1.Method = http.MethodPost
	req1.URL = testServer.URL
	req1.SetBody("ya.ru")
	resp1, _ := req1.Send()

	// wrong delete request
	req2 := resty.New().R()
	req2.Method = http.MethodDelete
	req2.URL = testServer.URL + url
	body2 := `["19xtf1u5", "19xtf1u5", "19xtf1tt"]`
	req2.SetHeader("Content-Type", "application/json")
	req2.SetBody(body2)
	req2.SetCookies(resp1.Cookies())
	resp2, err := req2.Send()
	assert.NoError(t, err, "error making HTTP request")
	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode(), "Response code didn't match expected")

	// valid delete request
	req3 := resty.New().R()
	req3.Method = http.MethodDelete
	req3.URL = testServer.URL + url
	body3 := `["19xtf1ts"]`
	req3.SetHeader("Content-Type", "application/json")
	req3.SetBody(body3)
	req3.SetCookies(resp1.Cookies())
	resp3, err := req3.Send()
	assert.NoError(t, err, "error making HTTP request")
	assert.Equal(t, http.StatusAccepted, resp3.StatusCode(), "Response code didn't match expected")

	// invalid empty delete request
	req4 := resty.New().R()
	req4.Method = http.MethodDelete
	req4.URL = testServer.URL + url
	body4 := `[]`
	req4.SetHeader("Content-Type", "application/json")
	req4.SetBody(body4)
	req4.SetCookies(resp1.Cookies())
	resp4, err := req4.Send()
	assert.NoError(t, err, "error making HTTP request")
	assert.Equal(t, http.StatusBadRequest, resp4.StatusCode(), "Response code didn't match expected")

	testServer.Close()
}

func TestServer_userURLsHandler(t *testing.T) {
	const url = "/api/user/urls"
	setup()

	// create URL request
	req1 := resty.New().R()
	req1.Method = http.MethodPost
	req1.URL = testServer.URL
	req1.SetBody("ya.ru")
	resp1, _ := req1.Send()

	// getting request
	req2 := resty.New().R()
	req2.Method = http.MethodGet
	req2.URL = testServer.URL + url
	req2.SetCookies(resp1.Cookies())
	resp2, err := req2.Send()
	assert.NoError(t, err, "error making HTTP request")
	assert.Equal(t, http.StatusOK, resp2.StatusCode(), "Response code didn't match expected")
	assert.JSONEq(
		t,
		fmt.Sprintf(`[{"short_url": "http://%s/19xtf1ts", "original_url": "ya.ru"}]`, config.BaseURL),
		string(resp2.Body()),
	)

	// cleaning
	testServer.Close()
	setup()

	// getting nothing
	req4 := resty.New().R()
	req4.Method = http.MethodGet
	req4.URL = testServer.URL + url
	req4.SetCookies(resp1.Cookies())
	// TODO:
	//resp4, err := req4.Send()
	//assert.NoError(t, err, "error making HTTP request")
	//assert.Equal(t, http.StatusNoContent, resp4.StatusCode(), "Response code didn't match expected")

	testServer.Close()
}

func TestServer_userURLsHandlerBadUserID(t *testing.T) {
	const url = "/api/user/urls"
	setup()

	// create URL request
	req := resty.New().R()
	req.Method = http.MethodPost
	req.URL = testServer.URL
	req.SetBody("ya.ru")

	resp, err := req.Send()
	t.Log("ERR", err)
	t.Log("RESP", resp)

}

func TestServer_pingHandler(t *testing.T) {
	const url = "/ping"
	config.FileStoragePath = "./storage_test.db"

	tests := []struct {
		name           string
		storageInst    repository.IStorage
		expectedStatus int
	}{
		{
			name:           "dbmaps",
			storageInst:    repository.NewDBMaps(),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "dbfiles",
			storageInst:    repository.NewDBFile(),
			expectedStatus: http.StatusInternalServerError,
		},
		//{
		//	name:           "dbfiles",
		//	storageInst:    storage.NewDBPgsql(),
		//	expectedStatus: http.StatusOK,
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()

			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL = testServer.URL + url
			res, _ := req.Send()

			assert.Equal(t, tt.expectedStatus, res.StatusCode())

			testServer.Close()
		})
	}

	_ = os.Remove(config.FileStoragePath)
}

func TestGzipCompression(t *testing.T) {
	// TODO: TestGzipCompression
	//handler := http.HandlerFunc(GzipMiddleware(httpServer.shortenHandler))
}

func TestServer_Start(t *testing.T) {
	setup()

	//t.Run("http ok", func(t *testing.T) {
	//	config.EnableHTTPS = false
	//	require.NotPanics(t, func() {
	//		serv.Start()
	//	})
	//})

	//t.Run("https ok", func(t *testing.T) {
	//	config.EnableHTTPS = true
	//	require.NotPanics(t, func() {
	//		serv.Start()
	//	})
	//})

	t.Run("http fatal", func(t *testing.T) {
		config.EnableHTTPS = false
		config.ServerAddress = "-"
		require.Panics(t, func() {
			testServer.Start()
		})
	})

	t.Run("https fatal", func(t *testing.T) {
		config.EnableHTTPS = true
		config.ServerAddress = "-"
		require.Panics(t, func() {
			testServer.Start()
		})
	})

}

func TestServer_Stop(t *testing.T) {
	setup()

	t.Run("normal stopping", func(t *testing.T) {
		go httpServer.Run()
		require.NotPanics(t, func() {
			httpServer.Stop()
		})
	})
}

func Test_getUserID(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/any",
		nil,
	)

	userID := int64(1)
	ctx := context.WithValue(req.Context(), secure.ContextUserIDKey, userID)
	req = req.WithContext(ctx)

	res, err := httpapi.GetUserID(req)
	require.NoError(t, err)
	require.Equal(t, userID, res)
}

func Test_getUserIDBadCtx(t *testing.T) {

	t.Run("wrong userID (qwer)", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/any",
			nil,
		)

		userID := "qwer"
		ctx := context.WithValue(req.Context(), secure.ContextUserIDKey, userID)
		req = req.WithContext(ctx)

		res, err := httpapi.GetUserID(req)
		//t.Log(err)
		require.ErrorContains(t, err, "invalid syntax")
		require.Equal(t, int64(0), res)
	})

	t.Run("wrong userID (0)", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/any",
			nil,
		)

		userID := 0
		ctx := context.WithValue(req.Context(), secure.ContextUserIDKey, userID)
		req = req.WithContext(ctx)

		res, err := httpapi.GetUserID(req)
		//t.Log(err)
		require.ErrorContains(t, err, "empty userID")
		require.Equal(t, int64(0), res)
	})
}
