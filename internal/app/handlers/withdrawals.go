package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MihailSergeenkov/gophermart/internal/app/common"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/MihailSergeenkov/gophermart/internal/app/services"
	"go.uber.org/zap"
)

func (h *Handlers) GetWithdrawals() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		withdrawals, err := h.services.GetWithdrawals(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error("failed to get withdrawals", zap.Error(err))
			return
		}

		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set(common.ContentTypeHeader, common.JSONContentType)
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		if err := enc.Encode(withdrawals); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error(common.EncRespErrStr, zap.Error(err))
			return
		}
	}
}
func (h *Handlers) AddWithdraw() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.AddWithdrawRequest

		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.logger.Error(common.ReadReqErrStr, zap.Error(err))
			return
		}

		err := h.services.AddWithdraw(r.Context(), req)
		if err != nil {
			if errors.Is(err, services.ErrOrderNumberValidation) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}

			if errors.Is(err, services.ErrInsufficientFunds) {
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error("failed to add withdraw", zap.Error(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
