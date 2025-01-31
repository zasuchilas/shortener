package grpcapi

import (
	"context"
	"errors"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/zasuchilas/shortener/internal/app/model"
	desc "github.com/zasuchilas/shortener/pkg/shortenergrpcv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteUserURLs _
func (i *Implementation) DeleteUserURLs(ctx context.Context, in *desc.DeleteUserURLsRequest) (*empty.Empty, error) {
	userID := int64(1)

	err := i.shortenerService.DeleteURLs(ctx, in.ShortUrls, userID)
	if err != nil {
		if errors.Is(err, model.ErrBadRequest) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}
