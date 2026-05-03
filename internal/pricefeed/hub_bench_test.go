package pricefeed

import (
	"context"
	"testing"
)

func BenchmarkHubPublish100(b *testing.B)   { benchHub(b, 100) }
func BenchmarkHubPublish1000(b *testing.B)  { benchHub(b, 1000) }
func BenchmarkHubPublish10000(b *testing.B) { benchHub(b, 10000) }

func benchHub(b *testing.B, subscribers int) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub()
	go hub.Run(ctx)

	for i := 0; i < subscribers; i++ {
		ch := hub.Subscribe()
		go func() {
			for range ch {
			}
		}()
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hub.Publish(PriceUpdate{Price: 150, Symbol: "BTCUSDT"})
	}
}
