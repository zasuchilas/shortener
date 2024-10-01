package models

import "time"

// URLRow is a row in file storage and postgresql storage
type URLRow struct {
	ID       int64  `json:"id"`
	ShortURL string `json:"short_url"`
	OrigURL  string `json:"original_url"`
	UserID   int64  `json:"user_id"`
	Deleted  bool   `json:"deleted"`
}

// UserRow is a row in secure data file
type UserRow struct {
	UserID   int64  `json:"user_id"`
	UserHash string `json:"user_hash"`
	UserDB   string `json:"user_db"`
}

// api/shorten

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

// /api/shorten/batch

type (
	ShortenBatchRequest  []ShortenBatchRequestItem
	ShortenBatchResponse []ShortenBatchResponseItem
)

type ShortenBatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// GET /api/user/urls

type UserURLsResponseItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type UserURLsResponse []UserURLsResponseItem

// DeleteTask is element for batch deleting chan
type DeleteTask struct {
	Time      time.Time
	UserID    int64
	ShortURLs []string
}
