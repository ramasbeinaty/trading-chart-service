package candlestick

import (
	"context"
	"sync"
	"time"

	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/base/logger"
	"go.uber.org/zap"
)

type CandlestickService struct {
	repo         IRepository
	lgr          logger.ILogger
	candlesticks map[string]*Candlestick
	mutex        sync.Mutex
}

func NewCandlestickService(
	repo IRepository,
	lgr logger.ILogger,
) *CandlestickService {
	return &CandlestickService{
		repo:         repo,
		lgr:          lgr,
		candlesticks: make(map[string]*Candlestick),
		mutex:        sync.Mutex{},
	}
}

func (c *CandlestickService) ProcessTicks(
	ctx context.Context,
	symbol string,
	price float64,
) error {
	lgr := c.lgr.Get(&ctx)
	lgr.Info("Processing ticks...")

	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := symbol + time.Now().Truncate(time.Minute).Format("200601021504")
	// update existing candlestick
	if candle, exists := c.candlesticks[key]; exists {
		lgr.Info(
			"Updating existing candlestick",
			zap.Any("candle", candle),
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
			Symbol:    symbol,
			Open:      price,
			High:      price,
			Low:       price,
			Close:     price,
			Timestamp: time.Now().Truncate(time.Minute),
		}
	}

	return nil
}

func (c *CandlestickService) CommitCompleteBars(
	ctx context.Context,
) error {
	lgr := c.lgr.Get(&ctx)
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
