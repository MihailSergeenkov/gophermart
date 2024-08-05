package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/MihailSergeenkov/gophermart/internal/app/common"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
)

var (
	ErrAnotherUserOrderExist = errors.New("order has another user")
	ErrUserOrderExist        = errors.New("user order already exist")
)

func (s *Services) AddOrder(ctx context.Context, number string) error {
	if err := checkOrderNumber(number); err != nil {
		return fmt.Errorf("failed check order number: %w", err)
	}

	order, isNewOrder, err := s.store.AddOrder(ctx, number)
	if err != nil {
		return fmt.Errorf("failed to add order: %w", err)
	}

	if isNewOrder {
		return nil
	}

	if order.UserID != ctx.Value(common.KeyUserID) {
		return ErrAnotherUserOrderExist
	}

	return ErrUserOrderExist
}

func (s *Services) GetOrders(ctx context.Context) ([]models.Order, error) {
	orders, err := s.store.GetOrdersByUserID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	return orders, nil
}
