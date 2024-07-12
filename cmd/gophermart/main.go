package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/MihailSergeenkov/gophermart/internal/app"
	"github.com/MihailSergeenkov/gophermart/internal/app/config"
	"github.com/MihailSergeenkov/gophermart/internal/app/data"
	"github.com/MihailSergeenkov/gophermart/internal/app/logger"
	"golang.org/x/sync/errgroup"
)

const (
	timeoutServerShutdown = time.Second * 5
	timeoutShutdown       = time.Second * 10
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	log.Println("bye-bye")
}

func run() error {
	rootCtx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	g, ctx := errgroup.WithContext(rootCtx)

	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	})

	c, err := config.Setup()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	l, err := logger.NewLogger(c.LogLevel)
	if err != nil {
		return fmt.Errorf("logger error: %w", err)
	}

	s, err := data.NewStorage(ctx, l, c)
	if err != nil {
		return fmt.Errorf("storage error: %w", err)
	}

	g.Go(func() error {
		<-ctx.Done()

		err := s.Close()
		if err != nil {
			return fmt.Errorf("failed to close db connection: %w", err)
		}

		return nil
	})

	app := app.InitApp(ctx, c, l, s)

	g.Go(func() (err error) {
		defer func() {
			errRec := recover()
			if errRec != nil {
				err = fmt.Errorf("a panic occurred: %v", errRec)
			}
		}()

		if err = app.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			return fmt.Errorf("listen and server has failed: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		defer log.Print("server has been shutdown")
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()

		if err := app.Shutdown(shutdownTimeoutCtx); err != nil {
			log.Printf("an error occurred during server shutdown: %v", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
