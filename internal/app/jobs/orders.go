package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MihailSergeenkov/gophermart/internal/app/clients"
	"github.com/MihailSergeenkov/gophermart/internal/app/data"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (bp *BackgroudProcessing) processOrdersAccrual(ctx context.Context) {
	var retryAfter time.Time

	ticker := time.NewTicker(bp.settings.ProcessOrderAccrualPeriod)
	stopGeneratorCh := make(chan struct{})
	defer close(stopGeneratorCh)
	stopWorkerCh := make(chan struct{}, bp.settings.ProcessOrderAccrualWorkers)
	defer close(stopWorkerCh)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if time.Now().Before(retryAfter) {
				continue
			}

			orders, err := bp.store.GetOrdersByStatus(ctx, "NEW", "PROCESSING")
			if err != nil {
				bp.logger.Error("failed to get orders from storage", zap.Error(err))
				continue
			}

			g := new(errgroup.Group)
			ordersCh := generator(stopGeneratorCh, orders)

			for range bp.settings.ProcessOrderAccrualWorkers {
				go worker(ctx, bp.store, bp.c.AccrualClient, g, stopWorkerCh, ordersCh)
			}

			if err := g.Wait(); err != nil {
				var pgxError *clients.TooManyRequestsError
				if errors.As(err, &pgxError) {
					retryAfter = pgxError.RetryAfter
					if _, opened := <-ordersCh; opened {
						stopGeneratorCh <- struct{}{}
					}

					for range bp.settings.ProcessOrderAccrualWorkers {
						stopWorkerCh <- struct{}{}
					}
				}
			}
		}
	}
}

func generator(stopCh <-chan struct{}, orders []models.Order) chan models.Order {
	inputCh := make(chan models.Order)

	go func() {
		defer close(inputCh)

		for _, order := range orders {
			select {
			case <-stopCh:
				return
			case inputCh <- order:
			}
		}
	}()

	return inputCh
}

func worker(ctx context.Context,
	s data.Storager,
	client *clients.AccrualClient,
	g *errgroup.Group,
	stopCh <-chan struct{},
	ordersCh <-chan models.Order) {
	for order := range ordersCh {
		select {
		case <-stopCh:
			return
		default:
			g.Go(func() error {
				err := processOrderAccrual(ctx, s, client, order)
				if err != nil && errors.Is(err, &clients.TooManyRequestsError{}) {
					return fmt.Errorf("too many requests: %w", err)
				}

				return nil
			})
		}
	}
}

func processOrderAccrual(ctx context.Context,
	s data.Storager,
	client *clients.AccrualClient,
	order models.Order) error {
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
