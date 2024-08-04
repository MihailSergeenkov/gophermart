package jobs

import (
	"context"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"go.uber.org/zap"
)

type BackgroudProcessing struct {
	settings *config.Settings
	logger   *zap.Logger
	store    Storager
}

type Storager interface {
	UpdateOrder(ctx context.Context, number string, status string, accrual float32) error
	GetOrdersByStatus(ctx context.Context, statuses ...string) ([]models.Order, error)
}

func NewBackgroudProcessing(settings *config.Settings, logger *zap.Logger, store Storager) *BackgroudProcessing {
	return &BackgroudProcessing{
		settings: settings,
		logger:   logger,
		store:    store,
	}
}

func (bp *BackgroudProcessing) Start(ctx context.Context) {
	go bp.processOrdersAccrual(ctx)
}
