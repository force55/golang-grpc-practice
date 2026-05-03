package redisbus

import (
	"context"
	"crypto-notifier/internal/pricefeed"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type Publisher struct {
	rdb     *redis.Client
	channel string
}

func NewPublisher(rdb *redis.Client, channel string) *Publisher {
	return &Publisher{rdb: rdb, channel: channel}
}

func (p *Publisher) Publish(ctx context.Context, update pricefeed.PriceUpdate) error {
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}

	return p.rdb.Publish(ctx, p.channel, data).Err()
}
