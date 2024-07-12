package clients

import (
	"fmt"
	"time"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"go.uber.org/zap"
)

type Clients struct {
	AccrualClient *AccrualClient
}

type TooManyRequestsError struct {
	RetryAfter time.Time
}

func (e *TooManyRequestsError) Error() string {
	return fmt.Sprintf("retry requests after %v", e.RetryAfter.Format("2006/01/02 15:04:05"))
}

func newToManyRequestsError(retryAfter int) error {
	return &TooManyRequestsError{
		RetryAfter: time.Now().Add(time.Duration(retryAfter) * time.Second),
	}
}

func InitClients(settings *config.Settings, logger *zap.Logger) *Clients {
	ac := newAccrualClient(settings, logger)

	return &Clients{
		AccrualClient: ac,
	}
}
