package binance

// TODO: A single connection to stream.binance.com is only valid for 24 hours
// 		Handle being disconnected at the 24 hour mark

const (
	BASE_ENDPOINT         = "stream.binance.com:9443"
	AGG_TRADE_STREAM_NAME = "aggTrade"
)
