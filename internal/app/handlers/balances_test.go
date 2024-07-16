package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MihailSergeenkov/gophermart/internal/app/common"
	"github.com/MihailSergeenkov/gophermart/internal/app/handlers/mocks"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetBalance(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	s := mocks.NewMockServicer(mockCtrl)
	l := mocks.NewMockLogger(mockCtrl)
	handlers := NewHandlers(s, l)

	errSome := errors.New("some error")

	type serviceResponse struct {
		res models.Balance
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
			name: "get balance success",
			serviceResponse: serviceResponse{
				res: models.Balance{
					Current:   100.22,
					Withdrawn: 100.22,
				},
				err: nil,
			},
			want: want{
				code:          http.StatusOK,
				contentType:   common.JSONContentType,
				body:          "{\"current\":100.22,\"withdrawn\":100.22}\n",
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "get balance failed",
			serviceResponse: serviceResponse{
				res: models.Balance{},
				err: errSome,
			},
			want: want{
				code:          http.StatusInternalServerError,
				contentType:   "",
				body:          "",
				errorLogTimes: 1,
				log:           "failed to get balance",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = s.EXPECT().GetBalance(gomock.Any()).Times(1).Return(test.serviceResponse.res, test.serviceResponse.err)
			_ = l.EXPECT().Error(test.want.log, zap.Error(errSome)).Times(test.want.errorLogTimes)

			request := httptest.NewRequest(http.MethodGet, "/api/user/balance", http.NoBody)
			w := httptest.NewRecorder()
			handlers.GetBalance()(w, request)

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
