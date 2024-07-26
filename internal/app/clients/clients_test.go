package clients

import (
	"testing"
	"time"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInitClients(t *testing.T) {
	t.Run("init clients success", func(t *testing.T) {
		accrualSystemAddress := "http://test.ru/qwerty"
		accrualRequestTimeout := time.Duration(1)
		settings := config.Settings{
			AccrualSystemAddress:  accrualSystemAddress,
			AccrualRequestTimeout: accrualRequestTimeout,
		}

		logger := zap.NewNop()
		c := InitClients(&settings, logger)

		assert.NotEmpty(t, c.AccrualClient)
		assert.NotEmpty(t, c.AccrualClient.client)
		assert.Equal(t, accrualSystemAddress, c.AccrualClient.systemAddress)
		assert.Equal(t, logger, c.AccrualClient.logger)
	})
}
