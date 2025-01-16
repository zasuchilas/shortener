package shortener_http_api_v1

import (
	"net/http"
	"time"
)

// ShortenerHTTPApiV1 _
type ShortenerHTTPApiV1 interface {
	PingHandler(http.ResponseWriter, *http.Request)
	ReadURLHandler(http.ResponseWriter, *http.Request)
	WriteURLHandler(http.ResponseWriter, *http.Request)
	ShortenHandler(http.ResponseWriter, *http.Request)
	ShortenBatchHandler(http.ResponseWriter, *http.Request)
	DeleteURLsHandler(http.ResponseWriter, *http.Request)
	UserURLsHandler(http.ResponseWriter, *http.Request)
	StatsHandler(http.ResponseWriter, *http.Request)
}

// POST api/shorten
type (
	ShortenRequest struct {
		URL string `json:"url"`
	}

	ShortenResponse struct {
		Result string `json:"result"`
	}
)

// POST /api/shorten/batch
type (
	// ShortenBatchRequestItem _
	ShortenBatchRequestItem struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	// ShortenBatchResponseItem _
	ShortenBatchResponseItem struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
)

// GET /api/user/urls
type (
	// UserURLsResponseItem _
	UserURLsResponseItem struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
)

// DeleteTask is element for batch deleting chan
type (
	DeleteTask struct {
		Time      time.Time
		UserID    int64
		ShortURLs []string
	}
)

// GET /api/internal/stats
type (
	// StatsResponse _
	StatsResponse struct {
		URLs  int `json:"urls"`
		Users int `json:"users"`
	}
)
