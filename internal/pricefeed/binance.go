package pricefeed

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"
	"sync"
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
	chunks := chunkSymbols(c.symbols, 1024)
	var wg sync.WaitGroup

	for _, chunk := range chunks {
		wg.Add(1)
		go func(symbols []string) {
			defer wg.Done()

			//reconnect if chunk disconnected
			for {
				err := c.connect(ctx, symbols)
				if ctx.Err() != nil {
					return
				}
				slog.Info("Chunk disconnected, reconnecting...", "error", err)
				time.Sleep(3 * time.Second)
			}
		}(chunk)
	}

	wg.Wait()
	return nil
}

func (c *BinanceClient) connect(ctx context.Context, symbols []string) error {

	baseUrl := "wss://stream.binance.com:9443/ws/"
	url := c.generateConnectUrl(baseUrl, symbols)
	slog.Info("Connecting to Binance", "url", url)
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

func (c *BinanceClient) generateConnectUrl(baseUrl string, symbols []string) string {
	//get each symbol and append @trade to it
	streams := make([]string, len(symbols))

	for i, s := range symbols {
		streams[i] = s + "@trade"
	}

	return baseUrl + strings.Join(streams, "/")
}

func chunkSymbols(symbols []string, chunkSize int) [][]string {
	chunks := make([][]string, 0)
	for i := 0; i < len(symbols); i += chunkSize {
		end := i + chunkSize
		if end > len(symbols) {
			end = len(symbols)
		}
		chunks = append(chunks, symbols[i:end])
	}
	return chunks
}
