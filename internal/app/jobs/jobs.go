package jobs

import (
	"context"

	"github.com/MihailSergeenkov/gophermart/internal/app/clients"
	"github.com/MihailSergeenkov/gophermart/internal/app/data"
	"go.uber.org/zap"
)

type BackgroudProcessing struct {
	c      *clients.Clients
	logger *zap.Logger
	store  data.Storager
}

func NewBackgroudProcessing(c *clients.Clients, logger *zap.Logger, store data.Storager) *BackgroudProcessing {
	return &BackgroudProcessing{
		c:      c,
		logger: logger,
		store:  store,
	}
}

func (bp *BackgroudProcessing) Start(ctx context.Context) {
	go bp.processOrdersAccrual(ctx)
}
