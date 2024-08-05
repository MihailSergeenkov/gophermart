package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/MihailSergeenkov/gophermart/internal/app/services"
	"go.uber.org/zap"
)

func (h *Handlers) AddOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			h.logger.Error(readReqErrStr, zap.Error(err))
			return
		}

		err = h.services.AddOrder(r.Context(), string(body))
		if err != nil {
			if errors.Is(err, services.ErrOrderNumberValidation) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}

			if errors.Is(err, services.ErrAnotherUserOrderExist) {
				w.WriteHeader(http.StatusConflict)
				return
			}

			if errors.Is(err, services.ErrUserOrderExist) {
				w.WriteHeader(http.StatusOK)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error("failed to add order", zap.Error(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *Handlers) GetOrders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orders, err := h.services.GetOrders(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error("failed to get orders", zap.Error(err))
			return
		}

		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set(ContentTypeHeader, JSONContentType)
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		if err := enc.Encode(orders); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error(encRespErrStr, zap.Error(err))
			return
		}
	}
}
