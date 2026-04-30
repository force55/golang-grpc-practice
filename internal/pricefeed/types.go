package pricefeed

type PriceUpdate struct {
	Symbol string
	Price  float64
	Time   int64
}

type BinanceTradeEvent struct {
	Symbol string `json:"s"`
	Price  string `json:"p"`
}
