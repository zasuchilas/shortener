package shortener

import (
	"context"
	"fmt"
	"time"
)

// Ping _
func (s *service) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := s.shortenerRepo.Ping(ctx); err != nil {
		return fmt.Errorf("postgresql is unavailable %w", err)
	}

	return nil
}
