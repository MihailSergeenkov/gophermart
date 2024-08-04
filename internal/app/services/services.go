package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
)

var ErrOrderNumberValidation = errors.New("order number has not been validated")

const (
	baseNumber = 10
	maxNumber  = 9
)

type Services struct {
	store    Storager
	settings *config.Settings
}

type Storager interface {
	GetUserByLogin(ctx context.Context, userLogin string) (models.User, error)
	AddUser(ctx context.Context, userLogin string, userPassword []byte) (models.User, error)
	GetOrdersByUserID(ctx context.Context) ([]models.Order, error)
	AddOrder(ctx context.Context, number string) (models.Order, bool, error)
	GetWithdrawals(ctx context.Context) ([]models.Withdraw, error)
	AddWithdraw(ctx context.Context, orderNumber string, sum float32) error
	GetBalance(ctx context.Context) (models.Balance, error)
	Ping(ctx context.Context) error
	Close() error
}

func NewServices(store Storager, settings *config.Settings) *Services {
	return &Services{
		store:    store,
		settings: settings,
	}
}

func checkOrderNumber(s string) error {
	number, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("failed to parse order number: %w", ErrOrderNumberValidation)
	}

	sum := 0

	for i := 1; number > 0; i++ {
		digit := number % baseNumber

		if i%2 == 0 {
			digit *= 2
			if digit > maxNumber {
				digit = digit%baseNumber + digit/baseNumber
			}
		}

		sum += digit
		number /= baseNumber
	}

	if sum%baseNumber != 0 {
		return fmt.Errorf("bad check sum for order number: %w", ErrOrderNumberValidation)
	}

	return nil
}
