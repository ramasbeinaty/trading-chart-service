package middlewares

import (
	"context"

	"github.com/ramasbeinaty/trading-chart-service/internal"
	"github.com/ramasbeinaty/trading-chart-service/pkg/app/middlewares/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func DefaultUnaryInterceptor(
	lgr *zap.Logger,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		// log incoming requests
		lgr.Info("Request received", zap.String("method", info.FullMethod))

		// check if ctx is cancelled before handling request
		if err := ctx.Err(); err != nil {
			lgr.Error(
				"Failed to handle request: context was cancelled",
				zap.Error(err),
			)
			return nil, status.Errorf(codes.Canceled, "request context was cancelled: %v", err)
		}

		// get trace context
		traceContext, err := tracing.GetTraceContext(ctx)
		if err != nil {
			lgr.Error("Failed to extract or generate trace context", zap.Error(err))
		}
		lgr.Info(
			"Request trace context received",
			zap.Any("traceContext", traceContext),
		)

		// add trace info to context
		ctx = context.WithValue(ctx, internal.TRACE_CONTEXT_KEY, traceContext)

		// proceed to handling request
		resp, err = handler(ctx, req)

		// handle error if any
		if err != nil {
			lgr.Error(
				"Error in completing the request",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)

			// convert to gRPC status
			if _, ok := status.FromError(err); !ok {
				// default to an internal error if err can't be converted
				err = status.Errorf(codes.Internal, "An internal error occurred")
				return nil, err
			}
		}

		lgr.Info(
			"Successfully completed the request",
			zap.String("method", info.FullMethod),
		)

		return resp, nil
	}
}

func RecoveryUnaryInterceptor(
	lgr *zap.Logger,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				lgr.Error(
					"Recovered from a panic",
					zap.Any("panic", r),
					zap.String("method", info.FullMethod),
				)
				err = status.Errorf(codes.Internal, "Internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
