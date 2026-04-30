package pricefeed

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"nhooyr.io/websocket"
)

type BinanceClient struct {
	hub     *Hub
	symbols []string
}

func NewBinanceClient(hub *Hub, symbols []string) *BinanceClient {
	return &BinanceClient{
		hub:     hub,
		symbols: symbols,
	}
}

func (c *BinanceClient) Run(ctx context.Context) error {
	for {
		err := c.connect(ctx)
		if ctx.Err() != nil {
			return nil
		}

		slog.Info("Binance connection closed, Reconnecting...", "error", err)
		time.Sleep(3 * time.Second) // Wait before reconnecting
	}
}

func (c *BinanceClient) connect(ctx context.Context) error {

	url := "wss://stream.binance.com:9443/ws/btcusdt@trade"
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return err
	}

	defer conn.Close(websocket.StatusNormalClosure, "")

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return err
		}

		//unmarshal data
		var event BinanceTradeEvent
		err = json.Unmarshal(data, &event)
		if err != nil {
			return err
		}

		//convert data price to float64
		price, err := strconv.ParseFloat(event.Price, 64)
		if err != nil {
			return err
		}

		c.hub.Publish(PriceUpdate{
			Symbol: event.Symbol,
			Price:  price,
			Time:   time.Now().Unix(),
		})

	}
}
