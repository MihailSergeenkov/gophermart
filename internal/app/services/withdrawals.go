package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/MihailSergeenkov/gophermart/internal/app/data"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
)

var ErrInsufficientFunds = errors.New("user has insufficient funds")

func (s *Services) GetWithdrawals(ctx context.Context) ([]models.Withdraw, error) {
	withdrawals, err := s.store.GetWithdrawals(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}

	return withdrawals, nil
}

func (s *Services) AddWithdraw(ctx context.Context, req models.AddWithdrawRequest) error {
	if err := checkOrderNumber(req.OrderNumber); err != nil {
		return fmt.Errorf("failed check order number: %w", err)
	}

	err := s.store.AddWithdraw(ctx, req.OrderNumber, req.Sum)
	if err != nil {
		if errors.Is(err, data.ErrUserInsufficientFunds) {
			return ErrInsufficientFunds
		}
		return fmt.Errorf("failed to add withdraw: %w", err)
	}

	return nil
}
