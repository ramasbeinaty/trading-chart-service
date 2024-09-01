package candlestickrepo

import (
	"context"
	"log"

	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/candlestick"
)

func (repo *_candlestickrepo) UpsertCandlestickBar(
	ctx context.Context,
	bar *candlestick.Candlestick,
) error {
	_, err := repo.db.Exec(
		queryUpsertCandlestickBar,
		bar.Symbol,
		bar.Open,
		bar.High,
		bar.Low,
		bar.Close,
	)
	if err != nil {
		log.Printf("Error: failed to upsert candlestickBar - %w", err)
		return err
	}

	return nil
}
