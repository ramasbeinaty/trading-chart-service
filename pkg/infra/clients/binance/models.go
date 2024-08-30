package binance

type TradeMessageDTO struct {
	EventType          string `json:"e"`
	EventTime          int64  `json:"E"`
	Symbol             string `json:"s"`
	AggTradeId         int64  `json:"a"`
	Price              string `json:"p"`
	Quantity           string `json:"q"`
	FirstTradeId       int64  `json:"f"`
	LastTradeId        int64  `json:"l"`
	TradeTime          int64  `json:"T"`
	IsBuyerMarketMaker bool   `json:"m"`
	Ignore             bool   `json:"M"`
}

type TradeMessageParsed struct {
	EventType          string  `json:"e"`
	EventTime          int64   `json:"E"`
	Symbol             string  `json:"s"`
	AggTradeId         int64   `json:"a"`
	Price              float64 `json:"p"`
	Quantity           float64 `json:"q"`
	FirstTradeId       int64   `json:"f"`
	LastTradeId        int64   `json:"l"`
	TradeTime          int64   `json:"T"`
	IsBuyerMarketMaker bool    `json:"m"`
	Ignore             bool    `json:"M"`
}
