package shortener

import (
	"context"

	"github.com/zasuchilas/shortener/internal/app/model"
)

func (s *service) Stats(ctx context.Context) (*model.Stats, error) {
	var (
		out model.Stats
		err error
	)

	out.URLs, err = s.shortenerRepo.Stats(ctx)
	if err != nil {
		return nil, err
	}

	out.Users = s.secure.UsersCount()

	return &out, nil
}
