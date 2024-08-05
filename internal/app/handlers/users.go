package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/MihailSergeenkov/gophermart/internal/app/services"
	"go.uber.org/zap"
)

func (h *Handlers) RegisterUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterUserRequest

		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.logger.Error(readReqErrStr, zap.Error(err))
			return
		}

		res, err := h.services.RegisterUser(r.Context(), req)

		if err != nil {
			if errors.Is(err, services.ErrUserValidationFields) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if errors.Is(err, services.ErrUserLoginExist) {
				w.WriteHeader(http.StatusConflict)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error("failed to register user", zap.Error(err))
			return
		}

		cookie := &http.Cookie{
			Name:     "AUTH_TOKEN",
			Value:    res.AuthToken,
			HttpOnly: true,
		}

		http.SetCookie(w, cookie)

		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handlers) LoginUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginUserRequest

		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.logger.Error(readReqErrStr, zap.Error(err))
			return
		}

		res, err := h.services.LoginUser(r.Context(), req)

		if err != nil {
			if errors.Is(err, services.ErrUserLoginCreds) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error("failed to login user", zap.Error(err))
			return
		}

		cookie := &http.Cookie{
			Name:     "AUTH_TOKEN",
			Value:    res.AuthToken,
			HttpOnly: true,
		}

		http.SetCookie(w, cookie)

		w.WriteHeader(http.StatusOK)
	}
}
