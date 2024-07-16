package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MihailSergeenkov/gophermart/internal/app/common"
	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/data/mocks"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAddOrder(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	currentUserID := 1
	ctx := context.WithValue(context.Background(), common.KeyUserID, currentUserID)
	errSome := errors.New("some error")

	type arg struct {
		number string
	}

	type mResponse struct {
		order      models.Order
		isNewOrder bool
		err        error
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
			name: "add order success",
			arg: arg{
				number: "12345678903",
			},
			mResponse: mResponse{
				order: models.Order{
					Number:     "12345678903",
					Status:     "NEW",
					Accrual:    0,
					UploadedAt: time.Now(),
					UserID:     currentUserID,
				},
				isNewOrder: true,
				err:        nil,
			},
			want: want{
				err: nil,
			},
		},
		{
			name: "add order failed",
			arg: arg{
				number: "12345678903",
			},
			mResponse: mResponse{
				order:      models.Order{},
				isNewOrder: false,
				err:        errSome,
			},
			want: want{
				err: errSome,
			},
		},
		{
			name: "another user order",
			arg: arg{
				number: "12345678903",
			},
			mResponse: mResponse{
				order: models.Order{
					Number:     "12345678903",
					Status:     "NEW",
					Accrual:    0,
					UploadedAt: time.Now(),
					UserID:     2,
				},
				isNewOrder: false,
				err:        nil,
			},
			want: want{
				err: ErrAnotherUserOrderExist,
			},
		},
		{
			name: "order already exist",
			arg: arg{
				number: "12345678903",
			},
			mResponse: mResponse{
				order: models.Order{
					Number:     "12345678903",
					Status:     "NEW",
					Accrual:    0,
					UploadedAt: time.Now(),
					UserID:     currentUserID,
				},
				isNewOrder: false,
				err:        nil,
			},
			want: want{
				err: ErrUserOrderExist,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = store.EXPECT().
				AddOrder(ctx, test.arg.number).
				Times(1).
				Return(test.mResponse.order, test.mResponse.isNewOrder, test.mResponse.err)

			err := s.AddOrder(ctx, test.arg.number)

			if test.mResponse.err != nil && assert.Error(t, err) {
				assert.ErrorContains(t, err, test.want.err.Error())
			}
		})
	}
}

func TestValidationFailedAddOrder(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	currentUserID := 1
	ctx := context.WithValue(context.Background(), common.KeyUserID, currentUserID)
	orderNumber := "123456789032222"

	t.Run("order number validation failed", func(t *testing.T) {
		_ = store.EXPECT().AddOrder(ctx, gomock.Any()).Times(0)

		err := s.AddOrder(ctx, orderNumber)

		if assert.Error(t, err) {
			assert.ErrorContains(t, err, ErrOrderNumberValidation.Error())
		}
	})
}

func TestSuccessGetOrders(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()
	orders := []models.Order{
		{
			Number:     "12345678",
			Status:     "NEW",
			Accrual:    0,
			UploadedAt: time.Now(),
			UserID:     1,
		},
	}

	_ = store.EXPECT().GetOrdersByUserID(ctx).Times(1).Return(orders, nil)

	t.Run("get orders success", func(t *testing.T) {
		result, err := s.GetOrders(ctx)
		assert.Equal(t, orders, result)
		assert.NoError(t, err)
	})
}

func TestFailedGetOrders(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()
	orders := []models.Order{}
	errSome := errors.New("some error")

	_ = store.EXPECT().GetOrdersByUserID(ctx).Times(1).Return(orders, errSome)

	t.Run("get orders failed", func(t *testing.T) {
		_, err := s.GetOrders(ctx)
		if assert.Error(t, err) {
			assert.ErrorContains(t, err, "failed to get orders", "some error")
		}
	})
}
