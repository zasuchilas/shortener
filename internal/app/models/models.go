package models

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

// URLRow is a row in file storage and postgresql storage
type URLRow struct {
	Uuid     int64  `json:"uuid"`
	ShortURL string `json:"short_url"`
	OrigURL  string `json:"original_url"`
}
