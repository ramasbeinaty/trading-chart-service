package candlestick

import (
	"time"
)

type Candlestick struct {
	Symbol    string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Timestamp time.Time
}
