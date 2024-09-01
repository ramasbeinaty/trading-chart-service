package logger

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type Logger struct {
	lgr       *zap.Logger
	lgrCtxKey string
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
		lgr:       lgr,
		lgrCtxKey: "lgrCtx",
	}, nil
}

func (l *Logger) Close() {
	l.lgr.Sync()
}

// Returns the logger associate with the context if available
// Else returns default logger
func (l *Logger) Get(
	ctx *context.Context,
) *zap.Logger {
	if ctx != nil {
		if lgr, ok := (*ctx).Value(l.lgrCtxKey).(*zap.Logger); ok {
			return lgr
		}
	}
	return l.lgr
}

// Returns the logger with w3c trace context (trace-id, parent-id, trace-flags)
// if trace context is of invalid format, returns default logger
// https://www.w3.org/TR/trace-context/#trace-id
func (l *Logger) GetWithTraceContext(
	traceContext string,
) *zap.Logger {
	trace := strings.Split(traceContext, "-")
	if len(trace) != 4 {
		return l.lgr
	}
	return l.lgr.With(
		zap.String("trace-id", trace[1]),
		zap.String("parent-id", trace[2]),
		zap.String("trace-flags", trace[3]),
	)
}
