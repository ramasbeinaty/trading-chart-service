package main

// https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#aggregate-trade-streams

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ramasbeinaty/trading-chart-service/pkg/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ========= Setup graceful system shutdown =========
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// ========= Start the app ========
	var wg sync.WaitGroup
	_app := app.StartAppService(ctx, &wg)

	// ========= Start GRPC server =========
	_grpc := app.StartGRPCServer(
		_app.Lgr,
		&wg,
		_app.CandlestickHandler,
	)

	// - Handle system shutdown
	<-quit
	log.Println("Shutting down system...")

	_app.StopAppService()
	_grpc.StopGrpcServer()
	wg.Wait()
	
	log.Println("Gracefully terminated the system, exiting...")
}
