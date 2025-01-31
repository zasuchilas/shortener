package converter

import (
	"github.com/zasuchilas/shortener/internal/app/model"
	"github.com/zasuchilas/shortener/pkg/shortenergrpcv1"
	"github.com/zasuchilas/shortener/pkg/shortenerhttpv1"
)

// ToShortenBatchInFromHTTP _
func ToShortenBatchInFromHTTP(in []shortenerhttpv1.ShortenBatchRequestItem) []model.ShortenBatchIn {
	result := make([]model.ShortenBatchIn, len(in))
	for i := range in {
		result[i] = model.ShortenBatchIn{
			CorrelationID: in[i].CorrelationID,
			OriginalURL:   in[i].OriginalURL,
		}
	}
	return result
}

// ToHTTPFromShortenBatchOut _
func ToHTTPFromShortenBatchOut(in []model.ShortenBatchOut) []shortenerhttpv1.ShortenBatchResponseItem {
	result := make([]shortenerhttpv1.ShortenBatchResponseItem, len(in))
	for i := range in {
		result[i] = shortenerhttpv1.ShortenBatchResponseItem{
			CorrelationID: in[i].CorrelationID,
			ShortURL:      in[i].ShortURL,
		}
	}
	return result
}

// ToShortenBatchInFromGRPC _
func ToShortenBatchInFromGRPC(in *shortenergrpcv1.ShortenBatchRequest) []model.ShortenBatchIn {
	result := make([]model.ShortenBatchIn, len(in.Items))
	for i := range in.Items {
		result[i] = model.ShortenBatchIn{
			CorrelationID: in.Items[i].CorrelationId,
			OriginalURL:   in.Items[i].OriginalUrl,
		}
	}
	return result
}

// ToGRPCFromShortenBatchOut _
func ToGRPCFromShortenBatchOut(in []model.ShortenBatchOut) *shortenergrpcv1.ShortenBatchResponse {
	result := shortenergrpcv1.ShortenBatchResponse{}
	for i := range in {
		result.Items[i] = &shortenergrpcv1.ShortenBatchResponse_Item{
			CorrelationId: in[i].CorrelationID,
			ShortUrl:      in[i].ShortURL,
		}
	}
	return &result
}
