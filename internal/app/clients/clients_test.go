package clients

import (
	"testing"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInitClients(t *testing.T) {
	t.Run("init clients success", func(t *testing.T) {
		settings := config.Settings{}
		logger := zap.NewNop()
		c := InitClients(&settings, logger)

		assert.NotEmpty(t, c.AccrualClient)
		assert.NotEmpty(t, c.AccrualClient.client)
		assert.Equal(t, &settings, c.AccrualClient.settings)
		assert.Equal(t, logger, c.AccrualClient.logger)
	})
}
