package pricefeed

import "testing"

func TestChunkSymbolsEmpty(t *testing.T) {
	chunks := chunkSymbols([]string{}, 2)

	if len(chunks) != 0 {
		t.Fatalf("Expected 0 chunks, got %d", len(chunks))
	}
}

func TestChunkSymbolsOneSymbol(t *testing.T) {
	chunks := chunkSymbols([]string{"BTCUSDT"}, 4)

	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}
}

func TestChunkSymbolsFiveSymbols(t *testing.T) {

	chunks := chunkSymbols([]string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "SOMEUSDT"}, 2)

	if len(chunks) != 3 {
		t.Fatalf("Expected 3 chunks, got %d", len(chunks))
	}
}

func TestGenerateConnectUrlOneSymbol(t *testing.T) {
	client := NewBinanceClient([]string{"BTCUSDT"}, nil)

	url := client.generateConnectUrl("wss://stream.binance.com:9443/ws/", client.symbols)

	if url != "wss://stream.binance.com:9443/ws/BTCUSDT@trade" {
		t.Fatalf("Expected url to be BTCUSDT@trade, got %s", url)
	}
}

func TestGenerateConnectUrlThreeSymbol(t *testing.T) {
	client := NewBinanceClient([]string{"BTCUSDT", "ETHUSDT", "ARAUSDT"}, nil)

	url := client.generateConnectUrl("wss://stream.binance.com:9443/ws/", client.symbols)

	if url != "wss://stream.binance.com:9443/ws/BTCUSDT@trade/ETHUSDT@trade/ARAUSDT@trade" {
		t.Fatalf("Expected url to be BTCUSDT@trade/ETHUSDT@trade/ARAUSDT@trade, got %s", url)
	}
}
