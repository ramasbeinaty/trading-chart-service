package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type tradeMessage struct {
	Price string `json:"p"`
}

type BinanceClient struct {
	DataChan chan<- float64
	stream   string
	symbol   string
	conn     *websocket.Conn
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewBinanceClient(
	dataChan chan<- float64,
	stream string,
	symbol string,
	ctx context.Context,
) *BinanceClient {
	ctx, cancel := context.WithCancel(ctx)
	return &BinanceClient{
		DataChan: dataChan,
		stream:   stream,
		symbol:   symbol,
		cancel:   cancel,
	}
}

func (bc *BinanceClient) ConnectToBinance() error {
	addr := fmt.Sprintf(
		"%s/ws/%s@%s",
		BASE_ENDPOINT,
		bc.symbol,
		bc.stream,
	)

	c, _, err := websocket.DefaultDialer.Dial("wss://"+addr, nil)
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
			bc.conn.WriteMessage(websocket.PongMessage, nil)
		case websocket.TextMessage:
			var msg tradeMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Println("Error unmarshaling message - %w", err)
				continue
			}

			price, err := strconv.ParseFloat(msg.Price, 64)
			if err != nil {
				log.Println("Error parsing price - %w", err)
				continue
			}

			bc.DataChan <- price
		default:
			log.Printf("Unhandled incoming message type, %d", messageType)
			continue
		}
	}
}
