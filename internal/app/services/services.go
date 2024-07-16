package services

import (
	"errors"
	"strconv"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/data"
)

var ErrOrderNumberValidation = errors.New("order number has not been validated")

const (
	baseNumber = 10
	maxNumber  = 9
)

type Services struct {
	store    data.Storager
	settings *config.Settings
}

func NewServices(store data.Storager, settings *config.Settings) *Services {
	return &Services{
		store:    store,
		settings: settings,
	}
}

func checkOrderNumber(s string) error {
	number, err := strconv.Atoi(s)
	if err != nil {
		return ErrOrderNumberValidation
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
		return ErrOrderNumberValidation
	}

	return nil
}
