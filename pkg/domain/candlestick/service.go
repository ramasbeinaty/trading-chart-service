package candlestick

import "github.com/ramasbeinaty/trading-chart-service/pkg/infra/repos"

type CandlestickService struct {
	Repository repos.CandlestickRepository
}

func (cs *CandlestickService) ProcessNewPrice(symbol string, price float64) error {
	candlestick, err := cs.Repository.GetLatestCandlestick(symbol)
	if err != nil {
		return err
	}

	candlestick.Update(price)
	return cs.Repository.Save(candlestick)
}
