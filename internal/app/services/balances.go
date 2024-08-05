package services

import (
	"context"
	"fmt"

	"github.com/MihailSergeenkov/gophermart/internal/app/models"
)

func (s *Services) GetBalance(ctx context.Context) (models.Balance, error) {
	balance, err := s.store.GetBalance(ctx)
	if err != nil {
		return models.Balance{}, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}
