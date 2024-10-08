package logger

import (
	"context"

	"go.uber.org/zap"
)

type ILogger interface {
	Get(
		ctx context.Context,
	) *zap.Logger
	Close()
}
