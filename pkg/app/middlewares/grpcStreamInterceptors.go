package middlewares

import (
	"context"

	"github.com/ramasbeinaty/trading-chart-service/internal"
	"github.com/ramasbeinaty/trading-chart-service/pkg/app/middlewares/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func DefaultStreamInterceptor(
	lgr *zap.Logger,
) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		// log incoming stream requests
		lgr.Info("Stream request received", zap.String("method", info.FullMethod))

		// check if ctx is cancelled before handling stream
		if err := ctx.Err(); err != nil {
			lgr.Error(
				"Failed to handle stream: context was cancelled",
				zap.Error(err),
			)
			return status.Errorf(codes.Canceled, "stream context was cancelled: %v", err)
		}

		// get trace context
		traceContext, err := tracing.GetTraceContext(ctx)
		if err != nil {
			lgr.Error("Failed to extract or generate trace context", zap.Error(err))
		}
		lgr.Info("Stream trace context received", zap.Any("traceContext", traceContext))

		// add trace info to context
		ctx = context.WithValue(ctx, internal.TRACE_CONTEXT_KEY, traceContext)

		// wrap stream with traced context
		streamWithContext := newStreamWithContext(ss, ctx)

		// proceed to handle stream
		err = handler(srv, streamWithContext)

		// handle error if any
		if err != nil {
			lgr.Error(
				"Error in completing the stream",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)

			// convert to gRPC status
			if _, ok := status.FromError(err); !ok {
				// default to an internal error if err can't be converted
				err = status.Errorf(codes.Internal, "An internal error occurred in the stream")
				return err
			}
		}

		lgr.Info(
			"Successfully completed the stream request",
			zap.String("method", info.FullMethod),
		)

		return err
	}
}

func RecoveryStreamInterceptor(
	lgr *zap.Logger,
) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		defer func() {
			if r := recover(); r != nil {
				lgr.Error(
					"Recovered from a panic in stream",
					zap.Any("panic", r),
					zap.String("method", info.FullMethod),
				)
				err = status.Errorf(codes.Internal, "Internal server error")
			}
		}()

		return handler(srv, ss)
	}
}

type streamWithContext struct {
	stream grpc.ServerStream
	ctx    context.Context
}

func newStreamWithContext(ss grpc.ServerStream, ctx context.Context) grpc.ServerStream {
	return &streamWithContext{stream: ss, ctx: ctx}
}

func (s *streamWithContext) Context() context.Context {
	return s.ctx
}

func (s *streamWithContext) RecvMsg(m interface{}) error {
	return s.stream.RecvMsg(m)
}

func (s *streamWithContext) SendMsg(m interface{}) error {
	return s.stream.SendMsg(m)
}

func (s *streamWithContext) SetHeader(md metadata.MD) error {
	return s.stream.SetHeader(md)
}

func (s *streamWithContext) SendHeader(md metadata.MD) error {
	return s.stream.SendHeader(md)
}

func (s *streamWithContext) SetTrailer(md metadata.MD) {
	s.stream.SetTrailer(md)
}
