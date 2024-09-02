package candlestick

import (
	"context"
	"sync"
	"time"

	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/base/logger"
	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/subscription"
	"github.com/ramasbeinaty/trading-chart-service/proto/candlestick/contracts"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CandlestickService struct {
	repo         IRepository
	lgr          logger.ILogger
	candlesticks map[string]*Candlestick
	mutex        sync.Mutex

	subscriptionService *subscription.SubscriptionService
}

func NewCandlestickService(
	repo IRepository,
	lgr logger.ILogger,
	subscriptionService *subscription.SubscriptionService,
) *CandlestickService {
	return &CandlestickService{
		repo:                repo,
		lgr:                 lgr,
		candlesticks:        make(map[string]*Candlestick),
		mutex:               sync.Mutex{},
		subscriptionService: subscriptionService,
	}
}

func (c *CandlestickService) ProcessTicks(
	ctx context.Context,
	symbol string,
	price float64,
	tradeTimestamp time.Time,
) error {
	lgr := c.lgr.Get(ctx)
	lgr.Info("Processing ticks...")

	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := symbol + tradeTimestamp.Truncate(time.Minute).Format("200602011504")
	var (
		candle *Candlestick
		exists bool
	)

	// update existing candlestick
	if candle, exists = c.candlesticks[key]; exists {
		lgr.Info(
			"Updating existing candlestick",
			zap.Any("candlestick", candle),
			zap.String("symbol", symbol),
			zap.Float64("price", price),
		)
		if price > candle.High {
			candle.High = price
		}
		if price < candle.Low {
			candle.Low = price
		}
		candle.Close = price
	} else {
		// create new candlestick
		lgr.Info(
			"Creating a new candlestick",
			zap.String("symbol", symbol),
			zap.Float64("price", price),
		)

		c.candlesticks[key] = &Candlestick{
			Symbol:         symbol,
			Open:           price,
			High:           price,
			Low:            price,
			Close:          price,
			TradeTimestamp: tradeTimestamp.Truncate(time.Minute),
		}

		candle = c.candlesticks[key]
	}

	c.subscriptionService.BroadcastToSubscribers(
		ctx,
		&contracts.Candlestick{
			Symbol:         candle.Symbol,
			OpenPrice:      candle.Open,
			HighPrice:      candle.High,
			LowPrice:       candle.Low,
			ClosePrice:     candle.Close,
			TradeTimestamp: timestamppb.New(candle.TradeTimestamp),
		},
	)

	return nil
}

func (c *CandlestickService) CommitCompleteBars(
	ctx context.Context,
) error {
	lgr := c.lgr.Get(ctx)
	lgr.Info("Committing complete bars...")

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key, candle := range c.candlesticks {
		err := c.repo.UpsertCandlestickBar(
			ctx,
			candle,
		)
		if err != nil {
			lgr.Error(
				"Error: failed to commit complete bar",
				zap.Any("candle", candle),
				zap.Error(err),
			)
			return err
		}

		// remove bar from memory after storing it in db
		delete(c.candlesticks, key)
	}

	lgr.Info("Successfully committed completed bars")

	return nil
}
