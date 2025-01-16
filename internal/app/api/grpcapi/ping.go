package grpcapi

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
)

func (i *Implementation) Ping(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	err := i.shortenerService.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
