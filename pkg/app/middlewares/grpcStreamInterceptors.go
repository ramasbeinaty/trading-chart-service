package middlewares

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func ContextStreamInterceptor(
	lgr *zap.Logger,
) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		// check if ctx is cancelled before handling stream
		if err := ctx.Err(); err != nil {
			lgr.Error(
				"Failed to handle stream: context was cancelled",
				zap.Error(err),
			)
			return err
		}

		// proceed to handle stream
		return handler(srv, ss)
	}
}
