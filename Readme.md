# TODO
- test grpc server
- update dockerfile
- update readme file with running instructions

# Start Here
## Run the app using the command (Change to dockerfile)
`go run .`

## You can use grpcurl to query grpc server
**To list the available methods**

`grpcurl -plaintext localhost:50051 list candlestick.CandlestickService
`

**To describe available methods**

`grpcurl -plaintext localhost:50051 describe candlestick.CandlestickService.SubscribeToCandlesticks
`

`grpcurl -plaintext localhost:50051 describe candlestick.CandlestickService.UnsubscribeFromCandlesticks
`

### Invoke available methods
For testing, set DEV_MODE = TRUE in docker-compose
The subscriber_id would then be set using a counter. So the first subscriber will have the id 1, second subscriber will have the id 2, etc.

- **SubscribeToCandlesticks**

`grpcurl -plaintext -d '{\"symbol\": \"BTCUSDT\"}' localhost:50051 candlestick.CandlestickService.SubscribeToCandlesticks
`

- **UnsubscribeFromCandlesticks**

`grpcurl -plaintext -d '{\"symbol\": \"BTCUSDT\", \"subscriber_id\": 1}' localhost:50051 candlestick.CandlestickService.UnsubscribeFromCandlesticks
`