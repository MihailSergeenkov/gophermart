package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MihailSergeenkov/gophermart/internal/app/handlers/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSuccessPing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	s := mocks.NewMockServicer(mockCtrl)
	l := mocks.NewMockLogger(mockCtrl)
	handlers := NewHandlers(s, l)

	_ = s.EXPECT().Ping(gomock.Any()).Times(1).Return(nil)
	_ = l.EXPECT().Error(gomock.Any(), gomock.Any()).Times(0)

	t.Run("ping success", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/ping", http.NoBody)
		w := httptest.NewRecorder()
		handlers.Ping()(w, request)

		res := w.Result()
		defer closeBody(t, res)

		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}

func TestFailedPing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	s := mocks.NewMockServicer(mockCtrl)
	l := mocks.NewMockLogger(mockCtrl)
	handlers := NewHandlers(s, l)

	errSome := errors.New("some error")
	_ = s.EXPECT().Ping(gomock.Any()).Times(1).Return(errSome)
	_ = l.EXPECT().Error("failed to connect to DB", zap.Error(errSome)).Times(1)

	t.Run("ping failed", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/ping", http.NoBody)
		w := httptest.NewRecorder()
		handlers.Ping()(w, request)

		res := w.Result()
		defer closeBody(t, res)

		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	})
}

func closeBody(t *testing.T, r *http.Response) {
	t.Helper()
	err := r.Body.Close()

	if err != nil {
		t.Log(err)
	}
}
