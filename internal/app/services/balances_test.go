package services

import (
	"context"
	"errors"
	"testing"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/MihailSergeenkov/gophermart/internal/app/services/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSuccessGetBalance(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()
	balance := models.Balance{
		Current:   0,
		Withdrawn: 0,
	}

	_ = store.EXPECT().GetBalance(ctx).Times(1).Return(balance, nil)

	t.Run("get balance success", func(t *testing.T) {
		result, err := s.GetBalance(ctx)
		assert.Equal(t, balance, result)
		assert.NoError(t, err)
	})
}

func TestFailedGetBalance(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()
	balance := models.Balance{}
	errSome := errors.New("some error")

	_ = store.EXPECT().GetBalance(ctx).Times(1).Return(balance, errSome)

	t.Run("get balance failed", func(t *testing.T) {
		_, err := s.GetBalance(ctx)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to get balance", "some error")
	})
}
