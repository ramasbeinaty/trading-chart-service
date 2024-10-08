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
	subscriberId int64,
	symbols []string,
	stream contracts.CandlestickService_SubscribeToCandlesticksServer,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	lgr := m.lgr.Get(ctx)
	lgr.Info(
		"Adding a subscriber",
		zap.Int64("subscriberId", subscriberId),
		zap.Strings("symbols", symbols),
		zap.Any("stream", stream),
	)

	sub, exists := m.GetSubscriber(subscriberId)
	if exists {
		lgr.Info("Updating existing subscriber")
		for _, s := range symbols {
			sub.Symbols[s] = true
		}
	} else {
		lgr.Info("Creating a new subscriber")

		_symbols := map[string]bool{}
		for _, s := range symbols {
			_symbols[s] = true
		}

		sub = &Subscriber{
			ID:      subscriberId,
			Symbols: _symbols,
			Stream:  stream,
		}

		m.subscribers[sub.ID] = sub
	}

	return nil
}

func (m *SubscriptionService) GetSubscriber(id int64) (*Subscriber, bool) {
	sub, exists := m.subscribers[id]
	return sub, exists
}

// if no symbol is provided, remove the subscriber and disconnect them from stream
// otherwise, just unsubscribe the subscriber from the symbol broadcast
func (m *SubscriptionService) RemoveSubscriber(
	ctx context.Context,
	subscriberId int64,
	symbols []string,
) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	lgr := m.lgr.Get(ctx)

	sub, exists := m.GetSubscriber(subscriberId)
	if !exists {
		lgr.Warn("subscriber not found, skipping...")
		return nil
	}

	if len(symbols) != 0 {
		for _, s := range symbols {
			delete(sub.Symbols, s)
		}

		// if not subscribed to any symbols, remove the subscriber and terminate stream
		if len(sub.Symbols) == 0 {
			delete(m.subscribers, subscriberId)
			sub.Stream.Context().Done()
		}
	} else {
		delete(m.subscribers, sub.ID)
		sub.Stream.Context().Done()
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
	lgr.Info(
		"Attempting to broadcast candlestick",
		zap.Any("candlestick", candlestick),
	)

	for _, sub := range m.subscribers {
		lgr.Debug("Checking subscriber", zap.Int64("ID", sub.ID), zap.Any("Symbols", sub.Symbols))

		if _, ok := sub.Symbols[candlestick.Symbol]; ok {
			lgr.Info("Found symbol in subscriber; sending", zap.String("Symbol", candlestick.Symbol), zap.Int64("SubscriberID", sub.ID))

			if err := sub.Stream.Send(candlestick); err != nil {
				lgr.Error(
					"Failed to send candlestick to subscriber. Connection might've broke",
					zap.Any("candlestick", candlestick),
					zap.Int64("subscriberId", sub.ID),
				)
				delete(m.subscribers, sub.ID)
				return fmt.Errorf("Failed to send candlestick to subscriber - %w", err)
			}
		} else {
			lgr.Info("Symbol not found for subscriber", zap.String("Symbol", candlestick.Symbol), zap.Int64("SubscriberID", sub.ID))
		}
	}

	return nil
}
