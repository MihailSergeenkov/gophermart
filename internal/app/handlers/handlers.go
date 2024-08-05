package handlers

import (
	"context"

	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"go.uber.org/zap/zapcore"
)

var (
	encRespErrStr     = "error encoding response"
	readReqErrStr     = "failed to read request body"
	ContentTypeHeader = "Content-Type"
	JSONContentType   = "application/json"
)

type Servicer interface {
	RegisterUser(ctx context.Context, req models.RegisterUserRequest) (models.RegisterUserResponse, error)
	LoginUser(ctx context.Context, req models.LoginUserRequest) (models.LoginUserResponse, error)
	AddOrder(ctx context.Context, number string) error
	GetOrders(ctx context.Context) ([]models.Order, error)
	GetWithdrawals(ctx context.Context) ([]models.Withdraw, error)
	AddWithdraw(ctx context.Context, req models.AddWithdrawRequest) error
	GetBalance(ctx context.Context) (models.Balance, error)
	Ping(ctx context.Context) error
}

type Logger interface {
	Error(msg string, fields ...zapcore.Field)
}

type Handlers struct {
	services Servicer
	logger   Logger
}

func NewHandlers(services Servicer, logger Logger) *Handlers {
	return &Handlers{
		services: services,
		logger:   logger,
	}
}
