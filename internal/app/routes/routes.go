package routes

import (
	"context"
	"net/http"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Handlerer interface {
	Ping() http.HandlerFunc
	RegisterUser() http.HandlerFunc
	LoginUser() http.HandlerFunc
	GetOrders() http.HandlerFunc
	GetWithdrawals() http.HandlerFunc
	AddOrder() http.HandlerFunc
	GetBalance() http.HandlerFunc
	AddWithdraw() http.HandlerFunc
}

type Storager interface {
	GetUserByID(ctx context.Context, userID int) (models.User, error)
}

var ContentTypeHeader = "Content-Type"
var JSONContentType = "application/json"

func NewRouter(h Handlerer, settings *config.Settings, l *zap.Logger, s Storager) chi.Router {
	r := chi.NewRouter()

	r.Get("/ping", h.Ping())

	r.Route("/api/user", func(r chi.Router) {
		r.Use(requestLogging(l))

		r.Group(func(r chi.Router) {
			r.Use(middleware.AllowContentType(JSONContentType))

			r.Post("/register", h.RegisterUser())
			r.Post("/login", h.LoginUser())
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware(settings, l, s))

			r.Group(func(r chi.Router) {
				r.Use(gzipMiddleware(l))

				r.Get("/orders", h.GetOrders())
				r.Get("/withdrawals", h.GetWithdrawals())
			})

			r.Post("/orders", h.AddOrder())

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", h.GetBalance())

				r.Group(func(r chi.Router) {
					r.Use(middleware.AllowContentType(JSONContentType))
					r.Post("/withdraw", h.AddWithdraw())
				})
			})
		})
	})

	return r
}
