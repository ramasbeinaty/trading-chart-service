package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type BinanceClient struct {
	TradeDataChan chan<- TradeMessageParsed
	stream        string
	symbols       []string
	conn          *websocket.Conn
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewBinanceClient(
	tradeDataChan chan<- TradeMessageParsed,
	stream string,
	symbols []string,
	ctx context.Context,
) *BinanceClient {
	ctx, cancel := context.WithCancel(ctx)

	return &BinanceClient{
		TradeDataChan: tradeDataChan,
		stream:        stream,
		symbols:       symbols,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (bc *BinanceClient) ConnectToBinance() error {
	// construct the stream path for one or more symbols
	streamQueries := make([]string, len(bc.symbols))
	for i, symbol := range bc.symbols {
		streamQueries[i] = fmt.Sprintf("%s@%s", symbol, bc.stream)
	}
	streamPath := strings.Join(streamQueries, "/")

	addr := fmt.Sprintf(
		"wss://%s/ws/%s",
		BASE_ENDPOINT,
		streamPath,
	)

	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Fatalf("Error dialing binance %s stream - %v", bc.stream, err)
		return err
	}

	bc.conn = c

	go bc.listen()
	return nil
}

func (bc *BinanceClient) reconnect() error {
	attempt := 0
	for {
		if err := bc.ConnectToBinance(); err == nil {
			log.Println("Successfully reconnected to binance")
			return nil
		}

		if attempt > 5 {
			log.Printf("Failed to reconnect to binance after %d attempts", attempt)
			return fmt.Errorf("Failed to reconnect to binance after many attempts")
		}

		attempt++
		backoff := time.Duration(min(attempt*2, 10)) * time.Second // capping backoff at 10 sec for simplicity
		time.Sleep(backoff)
	}
}

func (bc *BinanceClient) Close() error {
	bc.cancel()

	if bc.conn != nil {
		err := bc.conn.Close()
		if err != nil {
			log.Fatal("Failed to close binance connection - %w", err)
			return err
		}
	}

	return nil
}

func (bc *BinanceClient) listen() {
	defer bc.Close()

	for {
		// handle context cancellation or errors
		if bc.ctx.Err() != nil {
			log.Printf("Error with the context - %s", bc.ctx.Err().Error())
			return
		}

		// handle incoming messages from binance
		messageType, message, err := bc.conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message - %w", err)

			// binance connections disconnect after 24 hrs
			// network issues may occur
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("Connection closed unexpectedly - %v", err)
				bc.reconnect()
				continue
			}
			return
		}

		switch messageType {
		case websocket.PingMessage:
			log.Println("PONG!")
			bc.conn.WriteMessage(websocket.PongMessage, nil)
		case websocket.TextMessage:
			var msg TradeMessageDTO
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Println("Error unmarshaling message - %w", err)
				continue
			}

			price := 0.0
			if msg.Price != "" {
				price, err = strconv.ParseFloat(msg.Price, 64)
				if err != nil {
					log.Println("Error parsing price - %w", err)
					continue
				}
			}

			qty := 0.0
			if msg.Quantity != "" {
				qty, err = strconv.ParseFloat(msg.Quantity, 64)
				if err != nil {
					log.Println("Error parsing quantity - %w", err)
					continue
				}
			}

			parsedMsg := TradeMessageParsed{
				msg.EventType,
				msg.EventTime,
				msg.Symbol,
				msg.AggTradeId,
				price,
				qty,
				msg.FirstTradeId,
				msg.LastTradeId,
				msg.TradeTime,
				msg.IsBuyerMarketMaker,
				msg.Ignore,
			}

			bc.TradeDataChan <- parsedMsg
		default:
			log.Printf("Unhandled incoming message type, %d", messageType)
			continue
		}
	}
}
