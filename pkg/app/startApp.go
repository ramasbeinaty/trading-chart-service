package app

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/ramasbeinaty/trading-chart-service/internal"
	"github.com/ramasbeinaty/trading-chart-service/pkg/app/handlers"
	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/base/utils"
	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/candlestick"
	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/subscription"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/clients/binance"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/clients/snowflake"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/config"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/db"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/logger"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/repos/candlestickrepo"
	"go.uber.org/zap"
)

type App struct {
	Lgr                *zap.Logger
	DB                 *sql.DB
	BinanceClient      *binance.BinanceClient
	CandlestickHandler *handlers.CandlestickHandler
}

func StartAppService(
	ctx context.Context,
	wg *sync.WaitGroup,
) *App {
	// ========= Setup infra layer =========
	// env configs
	cfg := config.NewConfig()
	_binanceConfig := config.NewBinanceConfig(cfg)
	_dbConfig := config.NewDBConfig(cfg)
	_snowflakeConfig := config.NewSnowflakeConfig(cfg)

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

	// snowflake
	_snowflakeClient := snowflake.NewSnowflakeClient(ctx, _snowflakeConfig)

	// binance
	tradeDataChan := make(chan binance.TradeMessageParsed)
	_binanceClient := binance.NewBinanceClient(
		tradeDataChan,
		binance.AGG_TRADE_STREAM_NAME,
		internal.TRADE_SYMBOLS,
		ctx,
		_binanceConfig,
	)

	// ========= Setup repositories =========
	_candlestickrepo := candlestickrepo.NewCandlestickRepository(_db)

	// ========= Setup domain layer =========
	_subscriptionService := subscription.NewSubscriptionService(_lgrInstance)

	_candlestickService := candlestick.NewCandlestickService(
		_candlestickrepo,
		_lgrInstance,
		_subscriptionService,
	)

	// ========= Setup app layer =========
	_candlestickHandler := handlers.NewCandlestickHandler(
		_candlestickService,
		_subscriptionService,
		_snowflakeClient,
	)

	// ========= Start the app =========
	runAppService(
		ctx,
		_lgr,
		wg,
		&tradeDataChan,
		_binanceClient,
		_candlestickService,
	)

	return &App{
		_lgr,
		_db,
		_binanceClient,
		_candlestickHandler,
	}
}

func (a *App) StopAppService() error {
	a.Lgr.Info("Stopping app service...")

	if err := a.BinanceClient.Close(); err != nil {
		a.Lgr.Error("Failed to close binance client", zap.Error(err))
		return err
	}

	return nil
}

func runAppService(
	ctx context.Context,
	lgr *zap.Logger,
	wg *sync.WaitGroup,
	tradeDataChan *chan binance.TradeMessageParsed,
	binanceClient *binance.BinanceClient,
	candlestickService *candlestick.CandlestickService,
) {
	if tradeDataChan == nil {
		panic(fmt.Errorf("Failed to start app service - tradeDataChan is nil"))
	}

	// connect to binance
	if err := binanceClient.ConnectToBinance(); err != nil {
		panic(fmt.Sprintf("Failed to connect to Binance - %s", err.Error()))
	}

	// process candlestick ticks
	wg.Add(1)
	go func() {
		defer wg.Done()
		for trade := range *tradeDataChan {
			fmt.Printf("Received trade: %v\n", trade)
			err := candlestickService.ProcessTicks(
				ctx,
				trade.Symbol,
				trade.Price,
				utils.ConvertUnixMillisToTime(trade.TradeTime),
			)
			if err != nil {
				lgr.Error(
					"Failed to process ticks",
					zap.Any("trade", trade),
				)
			}
		}
	}()

	// start minute ticker to store complete candlestick bars every 1 minute
	startMinuteTicker(
		ctx,
		lgr,
		wg,
		candlestickService,
	)
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
