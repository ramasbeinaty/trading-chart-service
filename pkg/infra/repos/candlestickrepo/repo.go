package candlestickrepo

import (
	"database/sql"

	"github.com/ramasbeinaty/trading-chart-service/pkg/domain/candlestick"
)

type _candlestickrepo struct {
	db *sql.DB
}

var _ candlestick.IRepository = (*_candlestickrepo)(nil)

func NewCandlestickRepository(db *sql.DB) *_candlestickrepo {
	return &_candlestickrepo{
		db: db,
	}
}

// Queries
const (
	queryUpsertCandlestickBar = `
	INSERT INTO candlesticks (
		symbol, 
		open_price, 
		high_price, 
		low_price, 
		close_price, 
		timestamp,
		)
    VALUES (
		$1, 
		$2, 
		$3, 
		$4, 
		$5, 
		$6,
		)
    ON CONFLICT (symbol, timestamp) 
	DO UPDATE
    SET high_price = EXCLUDED.high_price,
    	low_price = EXCLUDED.low_price,
        close_price = EXCLUDED.close_price
	`
)
