package pricefeed

import "context"

type PriceUpdate struct {
	Symbol string
	Price  float64
	Time   int64
}

type BinanceTradeEvent struct {
	Symbol string `json:"s"`
	Price  string `json:"p"`
}

type PricePublisher interface {
	Publish(ctx context.Context, update PriceUpdate) error
}
