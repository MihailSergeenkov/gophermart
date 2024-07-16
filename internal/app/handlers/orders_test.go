package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MihailSergeenkov/gophermart/internal/app/common"
	"github.com/MihailSergeenkov/gophermart/internal/app/handlers/mocks"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/MihailSergeenkov/gophermart/internal/app/services"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAddOrder(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	s := mocks.NewMockServicer(mockCtrl)
	l := mocks.NewMockLogger(mockCtrl)
	handlers := NewHandlers(s, l)

	orderNumber := "12345678"

	type serviceResponse struct {
		err error
	}

	type want struct {
		code          int
		errorLogTimes int
		log           string
	}

	tests := []struct {
		name            string
		serviceResponse serviceResponse
		want            want
	}{
		{
			name: "add order success",
			serviceResponse: serviceResponse{
				err: nil,
			},
			want: want{
				code:          http.StatusAccepted,
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "add order failed with ErrOrderNumberValidation",
			serviceResponse: serviceResponse{
				err: services.ErrOrderNumberValidation,
			},
			want: want{
				code:          http.StatusUnprocessableEntity,
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "add order failed with ErrAnotherUserOrderExist",
			serviceResponse: serviceResponse{
				err: services.ErrAnotherUserOrderExist,
			},
			want: want{
				code:          http.StatusConflict,
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "add order failed with ErrUserOrderExist",
			serviceResponse: serviceResponse{
				err: services.ErrUserOrderExist,
			},
			want: want{
				code:          http.StatusOK,
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "add order failed with some error",
			serviceResponse: serviceResponse{
				err: errors.New("some error"),
			},
			want: want{
				code:          http.StatusInternalServerError,
				errorLogTimes: 1,
				log:           "failed to add order",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = s.EXPECT().AddOrder(gomock.Any(), orderNumber).Times(1).Return(test.serviceResponse.err)
			_ = l.EXPECT().Error(test.want.log, zap.Error(test.serviceResponse.err)).Times(test.want.errorLogTimes)

			request := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(orderNumber))
			w := httptest.NewRecorder()
			handlers.AddOrder()(w, request)

			res := w.Result()
			defer closeBody(t, res)

			assert.Equal(t, test.want.code, res.StatusCode)
		})
	}
}

func TestGetOrders(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	s := mocks.NewMockServicer(mockCtrl)
	l := mocks.NewMockLogger(mockCtrl)
	handlers := NewHandlers(s, l)

	errSome := errors.New("some error")
	uploadedAt := time.Now()

	type serviceResponse struct {
		res []models.Order
		err error
	}

	type want struct {
		code          int
		contentType   string
		body          string
		errorLogTimes int
		log           string
	}

	tests := []struct {
		name            string
		serviceResponse serviceResponse
		want            want
	}{
		{
			name: "get orders success",
			serviceResponse: serviceResponse{
				res: []models.Order{
					{
						Number:     "12345678",
						Status:     "NEW",
						Accrual:    0,
						UploadedAt: uploadedAt,
						UserID:     1,
					},
				},
				err: nil,
			},
			want: want{
				code:        http.StatusOK,
				contentType: common.JSONContentType,
				body: fmt.Sprintf(
					"[{\"number\":\"12345678\",\"status\":\"NEW\",\"uploaded_at\":%q}]\n",
					uploadedAt.Format(time.RFC3339Nano)),
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "get orders failed",
			serviceResponse: serviceResponse{
				res: []models.Order{},
				err: errSome,
			},
			want: want{
				code:          http.StatusInternalServerError,
				contentType:   "",
				body:          "",
				errorLogTimes: 1,
				log:           "failed to get orders",
			},
		},
		{
			name: "when no orders",
			serviceResponse: serviceResponse{
				res: []models.Order{},
				err: nil,
			},
			want: want{
				code:          http.StatusNoContent,
				contentType:   "",
				body:          "",
				errorLogTimes: 0,
				log:           "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = s.EXPECT().GetOrders(gomock.Any()).Times(1).Return(test.serviceResponse.res, test.serviceResponse.err)
			_ = l.EXPECT().Error(test.want.log, zap.Error(errSome)).Times(test.want.errorLogTimes)

			request := httptest.NewRequest(http.MethodGet, "/api/user/orders", http.NoBody)
			w := httptest.NewRecorder()
			handlers.GetOrders()(w, request)

			res := w.Result()
			defer closeBody(t, res)

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get(common.ContentTypeHeader))

			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, test.want.body, string(resBody))
		})
	}
}
