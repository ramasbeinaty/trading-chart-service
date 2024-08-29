package grpc

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ramasbeinaty/trading-chart-service/internal"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra"
	"go.uber.org/fx"
)

type app struct {
	inf *infra.Infrastructure
}

func NewApp(
	inf *infra.Infrastructure,
) *app {
	return &app{
		inf: inf,
	}
}

func (a *app) Start() {
	ctx := context.Background()

	app := fx.New(
		infra.DependencyOptions,
		fx.Invoke(registerHooks),
	)

	app.Start(ctx)
	defer app.Stop(context.Background())

	// - Setup graceful system shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// - Setup data migrations
	// db.RunMigrations(
	// 	ctx,
	// 	baseLgr,
	// 	a.dbctx,
	// 	db.GetMigrationScripts(),
	// )

	// - Set up middleware

	// - Start system processes
	var wg sync.WaitGroup

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	startGRPCServer()
	// }()

	for _, symbol := range internal.TRADE_SYMBOLS {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			binance.ConnectToBinance(
				s,
				binance.AGG_TRADE_STREAM_NAME,
				tradeDataChan)
		}(symbol)
	}

	// - Handle system shutdown
	<-quit
	log.Println("Shutting down system...")
}

func registerHooks() {
}
