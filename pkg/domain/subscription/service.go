package subscription

import (
	"context"
	"fmt"
	"sync"

	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/base/logger"
	"github.com/ramasbeinaty/trading-chart-service/proto/candlestick/contracts"
	"go.uber.org/zap"
)

type SubscriptionService struct {
	lgr         logger.ILogger
	mutex       sync.Mutex
	subscribers map[int64]*Subscriber // Keyed by subscriber ID
}

func NewSubscriptionService(
	lgr logger.ILogger,
) *SubscriptionService {
	return &SubscriptionService{
		lgr:         lgr,
		mutex:       sync.Mutex{},
		subscribers: make(map[int64]*Subscriber),
	}
}

func (m *SubscriptionService) AddUpdateSubscriber(
	ctx context.Context,
	cancel context.CancelFunc,
	subscriberId int64,
	symbol string,
	stream contracts.CandlestickService_SubscribeToCandlesticksServer,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	lgr := m.lgr.Get(ctx)
	lgr.Info(
		"Adding a subscriber",
		zap.Int64("subscriberId", subscriberId),
		zap.String("symbol", symbol),
		zap.Any("stream", stream),
	)

	sub, exists := m.GetSubscriber(subscriberId)
	if exists {
		lgr.Info("Updating existing subscriber")
		sub.Symbols[symbol] = true
	} else {
		lgr.Info("Creating a new subscriber")

		sub = &Subscriber{
			ID:      subscriberId,
			Symbols: map[string]bool{symbol: true},
			Stream:  stream,
			Cancel:  cancel,
		}

		m.subscribers[sub.ID] = sub
	}

	return nil
}

func (m *SubscriptionService) GetSubscriber(id int64) (*Subscriber, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sub, exists := m.subscribers[id]
	return sub, exists
}

// if symbol is nil, remove subscriber and disconnect them from stream
// otherwise, unsubscribe the subscriber from the symbol
func (m *SubscriptionService) RemoveSubscriber(
	ctx context.Context,
	subscriberId int64,
	symbol *string,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	lgr := m.lgr.Get(ctx)

	sub, exists := m.GetSubscriber(subscriberId)
	if !exists {
		lgr.Warn("subscriber not found, skipping...")
		return nil
	}

	if symbol != nil && *symbol != "" {
		delete(sub.Symbols, *symbol)

		// if not subscribed to any symbols, remove the subscriber and terminate stream
		if len(sub.Symbols) == 0 {
			delete(m.subscribers, subscriberId)
			sub.Cancel()
		}
	} else {
		delete(m.subscribers, sub.ID)
		sub.Cancel()
	}

	return nil
}

func (m *SubscriptionService) BroadcastToSubscribers(
	ctx context.Context,
	candlestick *contracts.Candlestick,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	lgr := m.lgr.Get(ctx)
	for _, sub := range m.subscribers {
		if _, ok := sub.Symbols[candlestick.Symbol]; ok {
			if err := sub.Stream.Send(candlestick); err != nil {
				lgr.Error(
					"Failed to send candlestick to subscriber. Connection might've broke",
					zap.Any("candlestick", candlestick),
					zap.Int64("subscriberId", sub.ID),
				)
				delete(m.subscribers, sub.ID)
				return fmt.Errorf("Failed to send candlestick to subscriber - %w", err)
			}
		}
	}

	return nil
}
