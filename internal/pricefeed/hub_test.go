package pricefeed

import (
	"context"
	"testing"
	"time"
)

func TestSubscribeAndPublish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//step 1 - prepare test
	hub := NewHub()
	go hub.Run(ctx)

	ch1 := hub.Subscribe()

	//step 2 - action
	go hub.Publish(PriceUpdate{Price: 150, Symbol: "BTC"})

	select {
	case update := <-ch1:
		if update.Price != 150 || update.Symbol != "BTC" {
			t.Errorf("Expected price 150 and symbol BTC, got %v", update)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Test timed out waiting for update on ch1")
	}

}

func TestUnsubscribe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub()
	go hub.Run(ctx)

	ch1 := hub.Subscribe()

	go hub.Unsubscribe(ch1)

	if _, ok := <-ch1; ok {
		t.Errorf("Expected ch1 to be closed, but it's still open")
	}

}

func TestMultipleSubscribers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub := NewHub()
	go hub.Run(ctx)

	ch1 := hub.Subscribe()
	ch2 := hub.Subscribe()

	go hub.Publish(PriceUpdate{Price: 150, Symbol: "BTC"})

	select {
	case update := <-ch1:
		if update.Price != 150 || update.Symbol != "BTC" {
			t.Errorf("Expected price 150 and symbol BTC, got %v", update)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Test timed out waiting for update on ch1")
	}

	select {
	case update := <-ch2:
		if update.Price != 150 || update.Symbol != "BTC" {
			t.Errorf("Expected price 150 and symbol BTC, got %v", update)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Test timed out waiting for update on ch2")
	}

}
