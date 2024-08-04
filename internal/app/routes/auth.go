package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/MihailSergeenkov/gophermart/internal/app/common"
	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/data"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func authMiddleware(settings *config.Settings, l *zap.Logger, s Storager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCookie, cookieErr := r.Cookie("AUTH_TOKEN")
			if cookieErr != nil {
				w.WriteHeader(http.StatusUnauthorized)
				l.Error("failed to fetch auth token", zap.Error(cookieErr))
				return
			}

			userID := getUserID(settings, authCookie.Value)

			if userID == -1 {
				w.WriteHeader(http.StatusUnauthorized)
				l.Error("failed to parse auth token")
				return
			}

			_, err := s.GetUserByID(r.Context(), userID)
			if err != nil {
				if errors.Is(err, data.ErrUserNotFound) {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				w.WriteHeader(http.StatusUnauthorized)
				l.Error("failed to get user from DB", zap.Error(err))
				return
			}

			newContext := context.WithValue(r.Context(), common.KeyUserID, userID)
			newRequest := r.WithContext(newContext)
			next.ServeHTTP(w, newRequest)
		})
	}
}

func getUserID(settings *config.Settings, tokenString string) int {
	claims := &models.Claims{}

	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(settings.SecretKey), nil
	})

	if err != nil {
		return -1
	}

	return claims.UserID
}
