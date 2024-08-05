package app

import (
	"context"
	"net/http"

	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/data"
	"github.com/MihailSergeenkov/gophermart/internal/app/handlers"
	"github.com/MihailSergeenkov/gophermart/internal/app/jobs"
	"github.com/MihailSergeenkov/gophermart/internal/app/routes"
	"github.com/MihailSergeenkov/gophermart/internal/app/services"
	"go.uber.org/zap"
)

func InitApp(ctx context.Context, settings *config.Settings, logger *zap.Logger, store *data.DBStorage) *http.Server {
	s := services.NewServices(store, settings)
	h := handlers.NewHandlers(s, logger)
	r := routes.NewRouter(h, settings, logger, store)
	j := jobs.NewBackgroudProcessing(settings, logger, store)
	j.Start(ctx)

	return &http.Server{
		Addr:    settings.RunAddr,
		Handler: r,
	}
}
