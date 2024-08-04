package services

import (
	"context"
	"errors"
	"testing"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/data"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/MihailSergeenkov/gophermart/internal/app/services/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()
	errSome := errors.New("some error")

	type arg struct {
		req models.RegisterUserRequest
	}

	type mResponse struct {
		user models.User
		err  error
	}

	type want struct {
		res models.RegisterUserResponse
		err error
	}

	tests := []struct {
		name      string
		arg       arg
		mResponse mResponse
		want      want
	}{
		{
			name: "register user success",
			arg: arg{
				req: models.RegisterUserRequest{
					Login:    "test",
					Password: "test",
				},
			},
			mResponse: mResponse{
				user: models.User{
					ID:       1,
					Login:    "test",
					Password: []byte("some_hash"),
				},
				err: nil,
			},
			want: want{
				res: models.RegisterUserResponse{
					AuthToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOjF9.06lbO3Sb1wyS45SCYsUxwrUyon5u6l1bnCbzwp83wbI",
				},
				err: nil,
			},
		},
		{
			name: "register user failed",
			arg: arg{
				req: models.RegisterUserRequest{
					Login:    "test",
					Password: "test",
				},
			},
			mResponse: mResponse{
				user: models.User{},
				err:  errSome,
			},
			want: want{
				res: models.RegisterUserResponse{},
				err: errSome,
			},
		},
		{
			name: "register user failed with constraint error",
			arg: arg{
				req: models.RegisterUserRequest{
					Login:    "test",
					Password: "test",
				},
			},
			mResponse: mResponse{
				user: models.User{},
				err:  &pgconn.PgError{Code: pgerrcode.UniqueViolation},
			},
			want: want{
				res: models.RegisterUserResponse{},
				err: ErrUserLoginExist,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = store.EXPECT().
				AddUser(ctx, test.arg.req.Login, gomock.Any()).
				Times(1).
				Return(test.mResponse.user, test.mResponse.err)

			result, err := s.RegisterUser(ctx, test.arg.req)

			assert.Equal(t, test.want.res, result)

			if test.mResponse.err != nil && assert.Error(t, err) {
				assert.ErrorContains(t, err, test.want.err.Error())
			}
		})
	}
}

func TestValidationFailedRegisterUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()

	type arg struct {
		req models.RegisterUserRequest
	}

	type want struct {
		err error
	}

	tests := []struct {
		name string
		arg  arg
		want want
	}{
		{
			name: "when login empty",
			arg: arg{
				req: models.RegisterUserRequest{
					Login:    "",
					Password: "test",
				},
			},
			want: want{
				err: ErrUserValidationFields,
			},
		},
		{
			name: "when password empty",
			arg: arg{
				req: models.RegisterUserRequest{
					Login:    "test",
					Password: "",
				},
			},
			want: want{
				err: ErrUserValidationFields,
			},
		},
		{
			name: "when password very big",
			arg: arg{
				req: models.RegisterUserRequest{
					Login:    "test",
					Password: "testtesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttest",
				},
			},
			want: want{
				err: bcrypt.ErrPasswordTooLong,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = store.EXPECT().AddUser(ctx, gomock.Any(), gomock.Any()).Times(0)

			_, err := s.RegisterUser(ctx, test.arg.req)

			if assert.Error(t, err) {
				assert.ErrorContains(t, err, test.want.err.Error())
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	store := mocks.NewMockStorager(mockCtrl)
	settings := config.Settings{}
	s := NewServices(store, &settings)

	ctx := context.Background()
	errSome := errors.New("some error")

	type arg struct {
		req models.LoginUserRequest
	}

	type mResponse struct {
		user models.User
		err  error
	}

	type want struct {
		res models.LoginUserResponse
		err error
	}

	tests := []struct {
		name      string
		arg       arg
		mResponse mResponse
		want      want
	}{
		{
			name: "login user success",
			arg: arg{
				req: models.LoginUserRequest{
					Login:    "test",
					Password: "test",
				},
			},
			mResponse: mResponse{
				user: models.User{
					ID:       1,
					Login:    "test",
					Password: []byte("$2a$10$eqoHdZljD4bk/zPKKGAPre6Mmq2mj8XxSrjF4SpavRy.pT/uxijYa"),
				},
				err: nil,
			},
			want: want{
				res: models.LoginUserResponse{
					AuthToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOjF9.06lbO3Sb1wyS45SCYsUxwrUyon5u6l1bnCbzwp83wbI",
				},
				err: nil,
			},
		},
		{
			name: "login user failed",
			arg: arg{
				req: models.LoginUserRequest{
					Login:    "test",
					Password: "test",
				},
			},
			mResponse: mResponse{
				user: models.User{},
				err:  errSome,
			},
			want: want{
				res: models.LoginUserResponse{},
				err: errSome,
			},
		},
		{
			name: "when user not found",
			arg: arg{
				req: models.LoginUserRequest{
					Login:    "test",
					Password: "test",
				},
			},
			mResponse: mResponse{
				user: models.User{},
				err:  data.ErrUserNotFound,
			},
			want: want{
				res: models.LoginUserResponse{},
				err: ErrUserLoginCreds,
			},
		},
		{
			name: "when incorrect password",
			arg: arg{
				req: models.LoginUserRequest{
					Login:    "test",
					Password: "test2",
				},
			},
			mResponse: mResponse{
				user: models.User{
					ID:       1,
					Login:    "test",
					Password: []byte("$2a$10$eqoHdZljD4bk/zPKKGAPre6Mmq2mj8XxSrjF4SpavRy.pT/uxijYa"),
				},
				err: nil,
			},
			want: want{
				res: models.LoginUserResponse{},
				err: ErrUserLoginCreds,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = store.EXPECT().GetUserByLogin(ctx, test.arg.req.Login).Times(1).Return(test.mResponse.user, test.mResponse.err)

			result, err := s.LoginUser(ctx, test.arg.req)

			assert.Equal(t, test.want.res, result)

			if test.mResponse.err != nil && assert.Error(t, err) {
				assert.ErrorContains(t, err, test.want.err.Error())
			}
		})
	}
}
