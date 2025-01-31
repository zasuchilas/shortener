package grpcapi

import (
	"context"
	desc "github.com/zasuchilas/shortener/pkg/shortenergrpcv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *Implementation) Shorten(ctx context.Context, in *desc.ShortenRequest) (*desc.ShortenResponse, error) {

	userID := int64(1)

	readyURL, conflict, err := i.shortenerService.WriteURL(ctx, in.Url, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if conflict {
		return nil, status.Error(codes.AlreadyExists, "conflict")
	}

	return &desc.ShortenResponse{Result: readyURL}, nil
}
