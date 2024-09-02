package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Logger struct {
	lgr *zap.Logger
}

func NewLogger() (*Logger, error) {
	lgr, err := zap.NewProduction(
		zap.AddCaller(),
	)
	if err != nil {
		return nil, err
	}
	if lgr == nil {
		return nil, fmt.Errorf("Failed to build a logger")
	}
	zap.ReplaceGlobals(lgr)
	return &Logger{
		lgr: lgr,
	}, nil
}

func (l *Logger) Close() {
	l.lgr.Sync()
}

func (l *Logger) Get(
	ctx context.Context,
) *zap.Logger {
	return l.lgr
}
