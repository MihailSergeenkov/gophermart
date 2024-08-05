package routes

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

const defaultStatus = 200

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size

	return size, err //nolint:wrapcheck // Нужно обернуть, но возврат должен остаться оригинальным
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func requestLogging(l *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			responseData := &responseData{
				status: defaultStatus,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			start := time.Now()
			uri := r.RequestURI
			method := r.Method

			next.ServeHTTP(&lw, r)

			duration := time.Since(start)

			l.Info("got incoming HTTP request",
				zap.String("uri", uri),
				zap.String("method", method),
				zap.String("duration", duration.String()),
				zap.Int("status", responseData.status),
				zap.Int("size", responseData.size),
			)
		})
	}
}
