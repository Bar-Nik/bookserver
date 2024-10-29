package logger

import (
	"context"
	"log/slog"
)

type key int

var myKey key = 5

func NewContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, myKey, logger)
}

func FromContext(ctx context.Context) (*slog.Logger, bool) {
	logger, ok := ctx.Value(myKey).(*slog.Logger)
	return logger, ok
}
