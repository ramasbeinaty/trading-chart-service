package candlestick

import "context"

type IRepository interface {
	UpsertCandlestickBar(
		ctx context.Context,
		bar *Candlestick,
	) error
}
