package grpcapi

import (
	"context"
	"github.com/zasuchilas/shortener/internal/app/converter"
	"github.com/zasuchilas/shortener/internal/app/logger"
	desc "github.com/zasuchilas/shortener/pkg/shortenergrpcv1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ShortenBatch _
func (i *Implementation) ShortenBatch(ctx context.Context, in *desc.ShortenBatchRequest) (*desc.ShortenBatchResponse, error) {

	userID := int64(1)

	out, err := i.shortenerService.ShortenBatch(ctx, converter.ToShortenBatchInFromGRPC(in), userID)
	if err != nil {
		logger.Log.Debug("shorten batch", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return converter.ToGRPCFromShortenBatchOut(out), nil
}
