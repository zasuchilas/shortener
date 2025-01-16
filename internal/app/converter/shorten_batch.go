package converter

import (
	"github.com/zasuchilas/shortener/internal/app/model"
	"github.com/zasuchilas/shortener/pkg/shortener_http_api_v1"
	"github.com/zasuchilas/shortener/pkg/shortener_v1"
)

func ToShortenBatchInFromHTTP(in []shortener_http_api_v1.ShortenBatchRequestItem) []model.ShortenBatchIn {
	result := make([]model.ShortenBatchIn, len(in))
	for i := range in {
		result[i] = model.ShortenBatchIn{
			CorrelationID: in[i].CorrelationID,
			OriginalURL:   in[i].OriginalURL,
		}
	}
	return result
}

func ToHTTPFromShortenBatchOut(in []model.ShortenBatchOut) []shortener_http_api_v1.ShortenBatchResponseItem {
	result := make([]shortener_http_api_v1.ShortenBatchResponseItem, len(in))
	for i := range in {
		result[i] = shortener_http_api_v1.ShortenBatchResponseItem{
			CorrelationID: in[i].CorrelationID,
			ShortURL:      in[i].ShortURL,
		}
	}
	return result
}

func ToShortenBatchInFromGRPC(in *shortener_v1.ShortenBatchRequest) []model.ShortenBatchIn {
	result := make([]model.ShortenBatchIn, len(in.Items))
	for i := range in.Items {
		result[i] = model.ShortenBatchIn{
			CorrelationID: in.Items[i].CorrelationId,
			OriginalURL:   in.Items[i].OriginalUrl,
		}
	}
	return result
}

func ToGRPCFromShortenBatchOut(in []model.ShortenBatchOut) *shortener_v1.ShortenBatchResponse {
	result := shortener_v1.ShortenBatchResponse{}
	for i := range in {
		result.Items[i] = &shortener_v1.ShortenBatchResponse_Item{
			CorrelationId: in[i].CorrelationID,
			ShortUrl:      in[i].ShortURL,
		}
	}
	return &result
}
