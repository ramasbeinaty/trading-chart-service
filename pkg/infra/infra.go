package infra

import (
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/clients/binance"
	"go.uber.org/fx"
)

var DependencyOptions = fx.Options(
	fx.Provide(
		binance.NewBinanceClient,
	),
)

type Infrastructure struct{}

func NewInfrastructure() *Infrastructure {
	return &Infrastructure{}
}

func (inf *Infrastructure) Start() {}

func (inf *Infrastructure) Stop() {}
