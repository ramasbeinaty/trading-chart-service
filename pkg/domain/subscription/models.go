package subscription

import (
	"context"

	"github.com/ramasbeinaty/trading-chart-service/proto/candlestick/contracts"
)

type Subscriber struct {
	ID      int64
	Symbols map[string]bool
	Stream  contracts.CandlestickService_SubscribeToCandlesticksServer
	Cancel  context.CancelFunc // to help terminate the stream
}
