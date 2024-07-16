package data

import (
	"context"
	"errors"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserInsufficientFunds = errors.New("user insufficient funds")
)

type Storager interface {
	GetUserByID(ctx context.Context, userID int) (models.User, error)
	GetUserByLogin(ctx context.Context, userLogin string) (models.User, error)
	AddUser(ctx context.Context, userLogin string, userPassword string) (models.User, error)
	GetOrdersByUserID(ctx context.Context) ([]models.Order, error)
	GetOrdersByStatus(ctx context.Context, statuses ...string) ([]models.Order, error)
	AddOrder(ctx context.Context, number string) (models.Order, bool, error)
	UpdateOrder(ctx context.Context, number string, status string, accrual float32) error
	GetWithdrawals(ctx context.Context) ([]models.Withdraw, error)
	AddWithdraw(ctx context.Context, orderNumber string, sum float32) error
	GetBalance(ctx context.Context) (models.Balance, error)
	Ping(ctx context.Context) error
	Close() error
}

func NewStorage(ctx context.Context, logger *zap.Logger, settings *config.Settings) (Storager, error) {
	return NewDBStorage(ctx, logger, settings.DatabaseURI)
}
