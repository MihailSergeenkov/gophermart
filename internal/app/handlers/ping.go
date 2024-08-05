package handlers

import (
	"net/http"

	"go.uber.org/zap"
)

func (h *Handlers) Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.services.Ping(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.logger.Error("failed to connect to DB", zap.Error(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
