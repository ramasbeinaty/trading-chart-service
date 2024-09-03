
# README for Trading Chart Service

## App Functionalities
- Reads tick data from binance data stream 
- Aggregates this data into OHLC Candlesticks with timeframe of 1 minute
- Serves a GRPC server
- Broadcasts the current symbol Candlestick bar to its subscribers
- Stores complete Candlestick bars in a Postgres database

## Start Here

### 1. Run the App
To run the app, use the command:
```bash
docker-compose up
```

### 2. Use grpcurl to Query the gRPC Server
**Note:** Below commands have been tested with bash. Might need to format for other terminals.

#### List Available Methods
```bash
grpcurl -plaintext localhost:50051 list candlestick.CandlestickService
```

#### Describe Available Methods
```bash
grpcurl -plaintext localhost:50051 describe candlestick.CandlestickService.SubscribeToCandlesticks

grpcurl -plaintext localhost:50051 describe candlestick.CandlestickService.UnsubscribeFromCandlesticks
```

### 3. Start Testing gRPC Server
**For testing**, set `ENV_ISDEVMODE=true` in `docker-compose.yaml`. The `subscriber_id` would then be set using a counter, meaning the first subscriber will have the ID 1, the second subscriber will have the id 2, etc.

#### SubscribeToCandlesticks
To subscribe to a single or multiple symbols

```bash
grpcurl -plaintext -d '{"symbols": ["BTCUSDT", "ETHUSDT", "PEPEUSDT"]}' localhost:50051 candlestick.CandlestickService.SubscribeToCandlesticks
```
#### UnsubscribeFromCandlesticks
To unsubscribe from specific symbol(s)
```bash
grpcurl -plaintext -d '{"symbols": ["BTCUSDT", "ETHUSDT"], "subscriber_id": 1}' localhost:50051 candlestick.CandlestickService.UnsubscribeFromCandlesticks
```

To unsubscribe from all symbols
```bash
grpcurl -plaintext -d '{"subscriber_id": 1}' localhost:50051 candlestick.CandlestickService.UnsubscribeFromCandlesticks
```
