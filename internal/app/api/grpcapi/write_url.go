package grpcapi

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	desc "github.com/zasuchilas/shortener/pkg/shortenergrpcv1"
)

// WriteURL _
func (i *Implementation) WriteURL(ctx context.Context, in *desc.WriteURLRequest) (*desc.WriteURLResponse, error) {

	userID := int64(1)

	readyURL, conflict, err := i.shortenerService.WriteURL(ctx, in.RawUrl, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if conflict {
		return nil, status.Errorf(codes.AlreadyExists, "url already exist %s", in.RawUrl)
	}

	return &desc.WriteURLResponse{
		ShortUrl: readyURL,
	}, nil
}
