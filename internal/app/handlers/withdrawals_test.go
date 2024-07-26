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

	"github.com/MihailSergeenkov/gophermart/internal/app/handlers/mocks"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/MihailSergeenkov/gophermart/internal/app/services"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAddWithdraw(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	s := mocks.NewMockServicer(mockCtrl)
	l := mocks.NewMockLogger(mockCtrl)
	handlers := NewHandlers(s, l)

	requestBody := "{\"order\":\"12345678\",\"sum\":100.22}"
	requestObject := models.AddWithdrawRequest{
		OrderNumber: "12345678",
		Sum:         100.22,
	}

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
				code:          http.StatusOK,
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
			name: "add order failed with ErrInsufficientFunds",
			serviceResponse: serviceResponse{
				err: services.ErrInsufficientFunds,
			},
			want: want{
				code:          http.StatusPaymentRequired,
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "add withdraw failed with some error",
			serviceResponse: serviceResponse{
				err: errors.New("some error"),
			},
			want: want{
				code:          http.StatusInternalServerError,
				errorLogTimes: 1,
				log:           "failed to add withdraw",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = s.EXPECT().AddWithdraw(gomock.Any(), requestObject).Times(1).Return(test.serviceResponse.err)
			_ = l.EXPECT().Error(test.want.log, zap.Error(test.serviceResponse.err)).Times(test.want.errorLogTimes)

			request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(requestBody))
			w := httptest.NewRecorder()
			handlers.AddWithdraw()(w, request)

			res := w.Result()
			defer closeBody(t, res)

			assert.Equal(t, test.want.code, res.StatusCode)
		})
	}
}

func TestFailedReadBodyAddWithdraw(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	s := mocks.NewMockServicer(mockCtrl)
	l := mocks.NewMockLogger(mockCtrl)
	handlers := NewHandlers(s, l)

	requestBody := "{\"order\":\"12345678\",\"sum\":100.22,adasd}"

	t.Run("failed to read request body", func(t *testing.T) {
		_ = s.EXPECT().AddWithdraw(gomock.Any(), gomock.Any()).Times(0)
		_ = l.EXPECT().Error("failed to read request body", gomock.Any()).Times(1)

		request := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", strings.NewReader(requestBody))
		w := httptest.NewRecorder()
		handlers.AddWithdraw()(w, request)

		res := w.Result()
		defer closeBody(t, res)

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestGetWithdrawals(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	s := mocks.NewMockServicer(mockCtrl)
	l := mocks.NewMockLogger(mockCtrl)
	handlers := NewHandlers(s, l)

	errSome := errors.New("some error")
	processedAt := time.Now()

	type serviceResponse struct {
		res []models.Withdraw
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
			name: "get withdrawals success",
			serviceResponse: serviceResponse{
				res: []models.Withdraw{
					{
						OrderNumber: "12345678",
						Sum:         100.22,
						ProcessedAt: processedAt,
					},
				},
				err: nil,
			},
			want: want{
				code:        http.StatusOK,
				contentType: JSONContentType,
				body: fmt.Sprintf(
					"[{\"processed_at\":%q,\"order\":\"12345678\",\"sum\":100.22}]\n",
					processedAt.Format(time.RFC3339Nano)),
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "get withdrawals failed",
			serviceResponse: serviceResponse{
				res: []models.Withdraw{},
				err: errSome,
			},
			want: want{
				code:          http.StatusInternalServerError,
				contentType:   "",
				body:          "",
				errorLogTimes: 1,
				log:           "failed to get withdrawals",
			},
		},
		{
			name: "when no withdrawals",
			serviceResponse: serviceResponse{
				res: []models.Withdraw{},
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
			_ = s.EXPECT().GetWithdrawals(gomock.Any()).Times(1).Return(test.serviceResponse.res, test.serviceResponse.err)
			_ = l.EXPECT().Error(test.want.log, zap.Error(errSome)).Times(test.want.errorLogTimes)

			request := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", http.NoBody)
			w := httptest.NewRecorder()
			handlers.GetWithdrawals()(w, request)

			res := w.Result()
			defer closeBody(t, res)

			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get(ContentTypeHeader))

			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, test.want.body, string(resBody))
		})
	}
}
