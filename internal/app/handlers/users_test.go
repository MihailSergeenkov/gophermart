package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MihailSergeenkov/gophermart/internal/app/handlers/mocks"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/MihailSergeenkov/gophermart/internal/app/services"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRegisterUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	s := mocks.NewMockServicer(mockCtrl)
	l := mocks.NewMockLogger(mockCtrl)
	handlers := NewHandlers(s, l)

	requestBody := "{\"login\":\"test\",\"password\":\"test\"}"
	requestObject := models.RegisterUserRequest{
		Login:    "test",
		Password: "test",
	}

	type serviceResponse struct {
		res models.RegisterUserResponse
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
			name: "register user success",
			serviceResponse: serviceResponse{
				res: models.RegisterUserResponse{
					AuthToken: "qwerty",
				},
				err: nil,
			},
			want: want{
				code:          http.StatusOK,
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "register user failed with ErrUserValidationFields",
			serviceResponse: serviceResponse{
				res: models.RegisterUserResponse{},
				err: services.ErrUserValidationFields,
			},
			want: want{
				code:          http.StatusBadRequest,
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "register user failed with ErrUserLoginExist",
			serviceResponse: serviceResponse{
				res: models.RegisterUserResponse{},
				err: services.ErrUserLoginExist,
			},
			want: want{
				code:          http.StatusConflict,
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "register user failed with some error",
			serviceResponse: serviceResponse{
				res: models.RegisterUserResponse{},
				err: errors.New("some error"),
			},
			want: want{
				code:          http.StatusInternalServerError,
				errorLogTimes: 1,
				log:           "failed to register user",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = s.EXPECT().
				RegisterUser(gomock.Any(), requestObject).
				Times(1).
				Return(test.serviceResponse.res, test.serviceResponse.err)

			_ = l.EXPECT().Error(test.want.log, zap.Error(test.serviceResponse.err)).Times(test.want.errorLogTimes)

			request := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(requestBody))
			w := httptest.NewRecorder()
			handlers.RegisterUser()(w, request)

			res := w.Result()
			defer closeBody(t, res)

			assert.Equal(t, test.want.code, res.StatusCode)

			if http.StatusOK == res.StatusCode {
				cookies := res.Cookies()
				var authCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == "AUTH_TOKEN" {
						authCookie = cookie
					}
				}
				assert.Equal(t, test.serviceResponse.res.AuthToken, authCookie.Value)
			}
		})
	}
}

func TestFailedReadBodyRegisterUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	s := mocks.NewMockServicer(mockCtrl)
	l := mocks.NewMockLogger(mockCtrl)
	handlers := NewHandlers(s, l)

	requestBody := "{\"login\":\"test\",\"password\":\"test\",adasd}"

	t.Run("failed to read request body", func(t *testing.T) {
		_ = s.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Times(0)
		_ = l.EXPECT().Error("failed to read request body", gomock.Any()).Times(1)

		request := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(requestBody))
		w := httptest.NewRecorder()
		handlers.RegisterUser()(w, request)

		res := w.Result()
		defer closeBody(t, res)

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestLoginUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	s := mocks.NewMockServicer(mockCtrl)
	l := mocks.NewMockLogger(mockCtrl)
	handlers := NewHandlers(s, l)

	requestBody := "{\"login\":\"test\",\"password\":\"test\"}"
	requestObject := models.LoginUserRequest{
		Login:    "test",
		Password: "test",
	}

	type serviceResponse struct {
		res models.LoginUserResponse
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
			name: "login user success",
			serviceResponse: serviceResponse{
				res: models.LoginUserResponse{
					AuthToken: "qwerty",
				},
				err: nil,
			},
			want: want{
				code:          http.StatusOK,
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "login user failed with ErrUserLoginCreds",
			serviceResponse: serviceResponse{
				res: models.LoginUserResponse{},
				err: services.ErrUserLoginCreds,
			},
			want: want{
				code:          http.StatusUnauthorized,
				errorLogTimes: 0,
				log:           "",
			},
		},
		{
			name: "login user failed with some error",
			serviceResponse: serviceResponse{
				res: models.LoginUserResponse{},
				err: errors.New("some error"),
			},
			want: want{
				code:          http.StatusInternalServerError,
				errorLogTimes: 1,
				log:           "failed to login user",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = s.EXPECT().
				LoginUser(gomock.Any(), requestObject).
				Times(1).
				Return(test.serviceResponse.res, test.serviceResponse.err)

			_ = l.EXPECT().Error(test.want.log, zap.Error(test.serviceResponse.err)).Times(test.want.errorLogTimes)

			request := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(requestBody))
			w := httptest.NewRecorder()
			handlers.LoginUser()(w, request)

			res := w.Result()
			defer closeBody(t, res)

			assert.Equal(t, test.want.code, res.StatusCode)

			if http.StatusOK == res.StatusCode {
				cookies := res.Cookies()
				var authCookie *http.Cookie
				for _, cookie := range cookies {
					if cookie.Name == "AUTH_TOKEN" {
						authCookie = cookie
					}
				}
				assert.Equal(t, test.serviceResponse.res.AuthToken, authCookie.Value)
			}
		})
	}
}

func TestFailedReadBodyLoginUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	s := mocks.NewMockServicer(mockCtrl)
	l := mocks.NewMockLogger(mockCtrl)
	handlers := NewHandlers(s, l)

	requestBody := "{\"login\":\"test\",\"password\":\"test\",adasd}"

	t.Run("failed to read request body", func(t *testing.T) {
		_ = s.EXPECT().LoginUser(gomock.Any(), gomock.Any()).Times(0)
		_ = l.EXPECT().Error("failed to read request body", gomock.Any()).Times(1)

		request := httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(requestBody))
		w := httptest.NewRecorder()
		handlers.LoginUser()(w, request)

		res := w.Result()
		defer closeBody(t, res)

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}
