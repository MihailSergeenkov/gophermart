package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MihailSergeenkov/gophermart/internal/app/clients"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"go.uber.org/zap"
)

func (bp *BackgroudProcessing) processOrdersAccrual(ctx context.Context) {
	var retryAfter time.Time
	client := clients.NewAccrualClient(&bp.settings.Accrual, bp.logger)

	ticker := time.NewTicker(bp.settings.ProcessOrderAccrualPeriod)

	ordersCh := make(chan models.Order)
	defer close(ordersCh)

	for range bp.settings.ProcessOrderAccrualWorkers {
		go worker(ctx, bp, client, ordersCh, &retryAfter)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ctx.Err() != nil {
				return
			}

			if time.Now().Before(retryAfter) {
				continue
			}

			orders, err := bp.store.GetOrdersByStatus(ctx, "NEW", "PROCESSING")
			if err != nil {
				bp.logger.Error("failed to get orders from storage", zap.Error(err))
				continue
			}

			go addOrdersToCh(ctx, ordersCh, orders)
		}
	}
}

func addOrdersToCh(ctx context.Context, ordersCh chan<- models.Order, orders []models.Order) {
	for _, order := range orders {
		select {
		case <-ctx.Done():
			return
		case ordersCh <- order:
		}
	}
}

func worker(
	ctx context.Context,
	bp *BackgroudProcessing,
	client *clients.AccrualClient,
	ordersCh <-chan models.Order,
	retryAfter *time.Time) {
	for order := range ordersCh {
		select {
		case <-ctx.Done():
			return
		default:
			now := time.Now()
			if now.Before(*retryAfter) {
				time.Sleep(retryAfter.Sub(now))
			}

			err := processOrderAccrual(ctx, bp.store, client, order)
			if err != nil {
				var pgxError *clients.TooManyRequestsError
				if errors.As(err, &pgxError) {
					*retryAfter = pgxError.RetryAfter
				}

				bp.logger.Error("failed process order accrual", zap.Error(err))
			}
		}
	}
}

func processOrderAccrual(ctx context.Context, s Storager, client *clients.AccrualClient, order models.Order) error {
	statusMap := map[string]string{
		"REGISTERED": "PROCESSING",
		"PROCESSING": "PROCESSING",
		"INVALID":    "INVALID",
		"PROCESSED":  "PROCESSED",
	}
	status, accrual, err := client.GetOrderAccrual(order.Number)
	if err != nil {
		return fmt.Errorf("failed process to get order accrual: %w", err)
	}

	err = s.UpdateOrder(ctx, order.Number, statusMap[status], accrual)
	if err != nil {
		return fmt.Errorf("failed process to update order: %w", err)
	}

	return nil
}
