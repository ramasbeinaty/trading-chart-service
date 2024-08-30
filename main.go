package main

// https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#aggregate-trade-streams

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ramasbeinaty/trading-chart-service/internal"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/clients/binance"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	// var wg sync.WaitGroup

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	startGRPCServer()
	// }()

	// - Setup infra clients
	// setup binance connection
	tradeDataChan := make(chan binance.TradeMessageParsed)
	binanceClient := binance.NewBinanceClient(
		tradeDataChan,
		binance.AGG_TRADE_STREAM_NAME,
		internal.TRADE_SYMBOLS,
		ctx,
	)
	if err := binanceClient.ConnectToBinance(); err != nil {
		panic(fmt.Sprintf("Failed to connect to Binance - %s", err.Error()))
	}

	go func() {
		for trade := range tradeDataChan {
			fmt.Printf("Received trade: %v\n", trade)
		}
	}()

	// - Handle system shutdown
	<-quit
	log.Println("Shutting down system...")

	binanceClient.Close()
	log.Println("Gracefully terminated the system, exiting...")
}
