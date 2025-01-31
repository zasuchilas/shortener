package grpcapi

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zasuchilas/shortener/internal/app/repository"
	desc "github.com/zasuchilas/shortener/pkg/shortenergrpcv1"
)

func (i *Implementation) ReadURL(ctx context.Context, in *desc.ReadURLRequest) (*desc.ReadURLResponse, error) {

	origURL, err := i.shortenerService.ReadURL(ctx, in.ShortUrl)
	if err != nil {
		if errors.Is(err, repository.ErrGone) {
			return nil, status.Errorf(codes.DataLoss, "%s is gone.", in.ShortUrl)
		}
		return nil, status.Error(codes.InvalidArgument, err.Error()) // according to the assignment, so, but postgresql may give an internal error
	}

	return &desc.ReadURLResponse{
		OrigUrl: origURL,
	}, nil
}
