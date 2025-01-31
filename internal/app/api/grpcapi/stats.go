package grpcapi

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/zasuchilas/shortener/internal/app/converter"
	desc "github.com/zasuchilas/shortener/pkg/shortenergrpcv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *Implementation) Stats(ctx context.Context, _ *empty.Empty) (*desc.StatsResponse, error) {

	out, err := i.shortenerService.Stats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return converter.ToGRPCFromStats(out), nil
}
