package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
	"go.uber.org/zap/zapcore"
)

type Settings struct {
	RunAddr              string        `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	AccrualSystemAddress string        `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8081"`
	DatabaseURI          string        `env:"DATABASE_URI" envDefault:"postgresql://postgres@localhost:5432/gophermart"`
	SecretKey            string        `env:"SECRET_KEY" envDefault:"1234567890"`
	LogLevel             zapcore.Level `env:"LOG_LEVEL" envDefault:"ERROR"`
}

func Setup() (*Settings, error) {
	s := Settings{LogLevel: zapcore.ErrorLevel}

	if err := s.parseFlags(); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &s, nil
}

func (s *Settings) parseFlags() error {
	err := env.Parse(s)

	if err != nil {
		return fmt.Errorf("env error: %w", err)
	}

	flag.StringVar(&s.RunAddr, "a", s.RunAddr, "address and port to run server")
	flag.StringVar(&s.AccrualSystemAddress, "r", s.AccrualSystemAddress, "address and port to accrual")
	flag.StringVar(&s.DatabaseURI, "d", s.DatabaseURI, "database URI")
	flag.StringVar(&s.SecretKey, "s", s.SecretKey, "secret key for generate auth token")
	flag.Func("l", `level for logger (default "ERROR")`, func(v string) error {
		lev, err := zapcore.ParseLevel(v)

		if err != nil {
			return fmt.Errorf("parse log level env error: %w", err)
		}

		s.LogLevel = lev
		return nil
	})

	flag.Parse()

	return nil
}
