package jobs

import (
	"context"

	"github.com/MihailSergeenkov/gophermart/internal/app/clients"
	"github.com/MihailSergeenkov/gophermart/internal/app/data"
	"go.uber.org/zap"
)

type BackgroudProcessing struct {
	clients *clients.Clients
	logger  *zap.Logger
	store   data.Storager
}

func NewBackgroudProcessing(clients *clients.Clients, logger *zap.Logger, store data.Storager) *BackgroudProcessing {
	return &BackgroudProcessing{
		clients: clients,
		logger:  logger,
		store:   store,
	}
}

func (bp *BackgroudProcessing) Start(ctx context.Context) {
	go bp.processOrdersAccrual(ctx)
}
