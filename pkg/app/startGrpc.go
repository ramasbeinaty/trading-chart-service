package app

import (
	"fmt"
	"net"
	"sync"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/ramasbeinaty/trading-chart-service/pkg/app/handlers"
	"github.com/ramasbeinaty/trading-chart-service/pkg/app/middlewares"
	candlestickpb "github.com/ramasbeinaty/trading-chart-service/proto/candlestick/contracts"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Grpc struct {
	server *grpc.Server
	lgr    *zap.Logger
}

func StartGRPCServer(
	lgr *zap.Logger,
	wg *sync.WaitGroup,
	candlestickHandler *handlers.CandlestickHandler,
) *Grpc {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				middlewares.RecoveryUnaryInterceptor(lgr),
				middlewares.DefaultUnaryInterceptor(lgr),
			),
		),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				middlewares.RecoveryStreamInterceptor(lgr),
				middlewares.DefaultStreamInterceptor(lgr)),
		),
	}
	s := grpc.NewServer(opts...)

	candlestickpb.RegisterCandlestickServiceServer(
		s,
		candlestickHandler,
	)

	// to query grpc server using grpcurl
	reflection.Register(s)

	wg.Add(1)
	go func() {
		defer wg.Done()

		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			panic(fmt.Errorf("failed to listen: %v", err))
		}

		lgr.Info(
			"gRPC server listening at",
			zap.Any("Address", lis.Addr()),
		)

		if err := s.Serve(lis); err != nil {
			panic(fmt.Errorf("failed to serve grpc server: %w", err))
		}
	}()

	return &Grpc{
		server: s,
	}
}

func (g *Grpc) StopGrpcServer() {
	g.lgr.Info("Stopping grpc server...")
	g.server.GracefulStop()
}
