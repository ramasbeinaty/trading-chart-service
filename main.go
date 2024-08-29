package main

// https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams#aggregate-trade-streams

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/ramasbeinaty/trading-chart-service/internal"
	"github.com/ramasbeinaty/trading-chart-service/pkg/infra/clients/binance"
)

func main() {
}

func connectToBinance() {
	var addr = "stream.binance.com:9443/ws/btcusdt@aggTrade"
	c, _, err := websocket.DefaultDialer.Dial("wss://"+addr, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
	}
}
