package candlestick

import (
	"sync"
	"time"
)

type Candlestick struct {
	Symbol    string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Timestamp time.Time
	mutex     sync.Mutex
}

func (c *Candlestick) Update(price float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Open == 0 {
		c.Open = price
	}
	if price > c.High {
		c.High = price
	}
	if price < c.Low || c.Low == 0 {
		c.Low = price
	}
	c.Close = price
}
