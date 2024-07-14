package services

import (
	"context"
	"errors"
	"testing"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/data/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSuccessPing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()

	_ = store.EXPECT().Ping(ctx).Times(1).Return(nil)

	t.Run("ping success", func(t *testing.T) {
		err := s.Ping(ctx)
		assert.NoError(t, err)
	})
}

func TestFailedPing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()
	errSome := errors.New("some error")

	_ = store.EXPECT().Ping(ctx).Times(1).Return(errSome)

	t.Run("ping failed", func(t *testing.T) {
		err := s.Ping(ctx)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to ping DB", "some error")
	})
}
