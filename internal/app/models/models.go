package models

// URLRow is a row in file storage and postgresql storage
type URLRow struct {
	UUID     int64  `json:"uuid"`
	ShortURL string `json:"short_url"`
	OrigURL  string `json:"original_url"`
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
