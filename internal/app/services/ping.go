package services

import (
	"context"
	"fmt"
)

func (s *Services) Ping(ctx context.Context) error {
	if err := s.store.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping DB %w", err)
	}

	return nil
}
