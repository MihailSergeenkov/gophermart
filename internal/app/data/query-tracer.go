package data

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type queryTracer struct {
	logger *zap.Logger
}

func (t *queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	t.logger.Info("Running query", zap.String("query", data.SQL), zap.Any("args", data.Args))
	return ctx
}

func (t *queryTracer) TraceQueryEnd(_ context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	t.logger.Info("End query", zap.Any("tag", data.CommandTag))
}
