// Package models is used to store service models.
package model

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

	// ShortenBatchIn is an item for shorten processing.
	ShortenBatchIn struct {
		CorrelationID string
		OriginalURL   string
	}

	// ShortenBatchOut is an item for shorten result.
	ShortenBatchOut struct {
		CorrelationID string
		ShortURL      string
	}

	// UserURL _
	UserURL struct {
		ShortURL    string
		OriginalURL string
	}

	// DeleteTask is element for batch deleting chan.
	DeleteTask struct {
		Time      time.Time
		UserID    int64
		ShortURLs []string
	}

	// Stats _
	Stats struct {
		URLs  int
		Users int
	}
)
