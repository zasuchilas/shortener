// Package models is used to store service models.
package models

import "time"

// Types
type (
	// URLRow is a row in file storage and postgresql storage
	URLRow struct {
		ID       int64  `json:"id"`
		ShortURL string `json:"short_url"`
		OrigURL  string `json:"original_url"`
		UserID   int64  `json:"user_id"`
		Deleted  bool   `json:"deleted"`
	}

	// UserRow is a row in secure data file
	UserRow struct {
		UserID   int64  `json:"user_id"`
		UserHash string `json:"user_hash"`
		UserDB   string `json:"user_db"`
	}
)

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
	ShortenBatchRequest []ShortenBatchRequestItem

	ShortenBatchResponse []ShortenBatchResponseItem

	ShortenBatchRequestItem struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	ShortenBatchResponseItem struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
)

// GET /api/user/urls
type (
	UserURLsResponseItem struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	UserURLsResponse []UserURLsResponseItem
)

// DeleteTask is element for batch deleting chan
type (
	DeleteTask struct {
		Time      time.Time
		UserID    int64
		ShortURLs []string
	}
)
