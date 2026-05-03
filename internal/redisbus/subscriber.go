package redisbus

import (
	"context"
	"crypto-notifier/internal/pricefeed"
	"encoding/json"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Subscriber struct {
	rdb     *redis.Client
	hub     *pricefeed.Hub
	channel string
}

func NewSubscriber(rdb *redis.Client, hub *pricefeed.Hub, channel string) *Subscriber {
	return &Subscriber{rdb: rdb, hub: hub, channel: channel}
}

func (s *Subscriber) Run(ctx context.Context) error {
	pubsub := s.rdb.Subscribe(ctx, s.channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}

			var update pricefeed.PriceUpdate
			if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
				slog.Error("Failed to unmarshal price update", "error", err)
				continue
			}

			s.hub.Publish(update)

		case <-ctx.Done():
			return nil
		}
	}

}
