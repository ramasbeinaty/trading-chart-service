package main

// https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#aggregate-trade-streams

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ramasbeinaty/trading-chart-service/internal"
	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/base/utils"
	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/candlestick"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/clients/binance"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/config"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/db"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/logger"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/repos/candlestickrepo"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// - Setup graceful system shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// - Set up middleware

	// - Start system processes
	var wg sync.WaitGroup

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	startGRPCServer()
	// }()

	// - Setup infra layer
	// env configs
	cfg := config.NewConfig()
	_binanceConfig := config.NewBinanceConfig(cfg)
	_dbConfig := config.NewDBConfig(cfg)

	// logger
	_lgrInstance, err := logger.NewLogger()
	if err != nil {
		panic("Error: Failed to initialize logger")
	}
	_lgr := _lgrInstance.Get(nil)

	// db
	_db, err := db.InitializeDB(_dbConfig)
	if err != nil {
		panic("Error: Failed to connect to db")
	}
	db.RunMigrations(
		ctx,
		_lgr,
		_db,
		db.GetMigrationScripts(),
	)

	// setup binance connection
	tradeDataChan := make(chan binance.TradeMessageParsed)
	binanceClient := binance.NewBinanceClient(
		tradeDataChan,
		binance.AGG_TRADE_STREAM_NAME,
		internal.TRADE_SYMBOLS,
		ctx,
		_binanceConfig,
	)
	if err := binanceClient.ConnectToBinance(); err != nil {
		panic(fmt.Sprintf("Failed to connect to Binance - %s", err.Error()))
	}

	// setup repositories
	_candlestickrepo := candlestickrepo.NewCandlestickRepository(_db)

	// - Setup domain layer
	// setup candlestick service
	_candlestickService := candlestick.NewCandlestickService(
		_candlestickrepo,
		_lgrInstance,
	)

	// process candlestick ticks
	wg.Add(1)
	go func() {
		defer wg.Done()
		for trade := range tradeDataChan {
			fmt.Printf("Received trade: %v\n", trade)
			err := _candlestickService.ProcessTicks(
				ctx,
				trade.Symbol,
				trade.Price,
				utils.ConvertUnixMillisToTime(trade.TradeTime),
			)
			if err != nil {
				_lgr.Error(
					"Failed to process ticks",
					zap.Any("trade", trade),
				)
			}
		}
	}()

	// start minute ticker to store complete candlestick bars every 1 minute
	startMinuteTicker(
		ctx,
		_lgr,
		&wg,
		_candlestickService,
	)

	// - Handle system shutdown
	<-quit
	log.Println("Shutting down system...")

	binanceClient.Close()
	log.Println("Gracefully terminated the system, exiting...")
}

func startMinuteTicker(
	ctx context.Context,
	lgr *zap.Logger,
	wg *sync.WaitGroup,
	candlestickService *candlestick.CandlestickService,
) {
	now := time.Now().UTC()

	delay := time.Minute - time.Duration(now.Second())*time.Second -
		time.Duration(now.Nanosecond())

	time.AfterFunc(delay, func() {
		ticker := time.NewTicker(time.Minute)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range ticker.C {
				err := candlestickService.CommitCompleteBars(ctx)
				if err != nil {
					lgr.Error(
						"Error: failed to commit complete bars",
						zap.Error(err),
					)
				}
			}
		}()
	},
	)
}
