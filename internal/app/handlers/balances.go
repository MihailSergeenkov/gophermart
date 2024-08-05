package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func (h *Handlers) GetBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := h.services.GetBalance(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error("failed to get balance", zap.Error(err))
			return
		}

		w.Header().Set(ContentTypeHeader, JSONContentType)
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		if err := enc.Encode(res); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error(encRespErrStr, zap.Error(err))
			return
		}
	}
}
