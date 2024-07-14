package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/data"
	"github.com/MihailSergeenkov/gophermart/internal/app/data/mocks"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAddWithdraw(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()
	errSome := errors.New("some error")

	type arg struct {
		req models.AddWithdrawRequest
	}

	type mResponse struct {
		err error
	}

	type want struct {
		err error
	}

	tests := []struct {
		name      string
		arg       arg
		mResponse mResponse
		want      want
	}{
		{
			name: "add withdraw success",
			arg: arg{
				req: models.AddWithdrawRequest{
					OrderNumber: "12345678903",
					Sum:         100,
				},
			},
			mResponse: mResponse{
				err: nil,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "add withdraw failed",
			arg: arg{
				req: models.AddWithdrawRequest{
					OrderNumber: "12345678903",
					Sum:         100,
				},
			},
			mResponse: mResponse{
				err: errSome,
			},
			want: want{
				err: errSome,
			},
		},
		{
			name: "user insufficient funds",
			arg: arg{
				req: models.AddWithdrawRequest{
					OrderNumber: "12345678903",
					Sum:         100,
				},
			},
			mResponse: mResponse{
				err: data.ErrUserInsufficientFunds,
			},
			want: want{
				err: ErrInsufficientFunds,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = store.EXPECT().AddWithdraw(ctx, test.arg.req.OrderNumber, test.arg.req.Sum).Times(1).Return(test.mResponse.err)

			err := s.AddWithdraw(ctx, test.arg.req)

			if test.mResponse.err != nil && assert.Error(t, err) {
				assert.ErrorContains(t, err, test.want.err.Error())
			}

		})
	}
}

func TestValidationFailedAddWithdraw(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()
	req := models.AddWithdrawRequest{
		OrderNumber: "123456789032222",
		Sum:         100,
	}

	t.Run("order number validation failed", func(t *testing.T) {
		_ = store.EXPECT().AddWithdraw(ctx, gomock.Any(), gomock.Any()).Times(0)

		err := s.AddWithdraw(ctx, req)

		if assert.Error(t, err) {
			assert.ErrorContains(t, err, ErrOrderNumberValidation.Error())
		}

	})
}

func TestSuccessGetWithdrawals(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()
	withdrawals := []models.Withdraw{
		{
			OrderNumber: "12345678",
			Sum:         0,
			ProcessedAt: time.Now(),
		},
	}

	_ = store.EXPECT().GetWithdrawals(ctx).Times(1).Return(withdrawals, nil)

	t.Run("get withdrawals success", func(t *testing.T) {
		result, err := s.GetWithdrawals(ctx)
		assert.Equal(t, withdrawals, result)
		assert.NoError(t, err)
	})
}

func TestFailedGetWithdrawals(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()
	withdrawals := []models.Withdraw{}
	errSome := errors.New("some error")

	_ = store.EXPECT().GetWithdrawals(ctx).Times(1).Return(withdrawals, errSome)

	t.Run("get withdrawals failed", func(t *testing.T) {
		_, err := s.GetWithdrawals(ctx)
		if assert.Error(t, err) {
			assert.ErrorContains(t, err, "failed to get withdrawals", "some error")
		}
	})
}
