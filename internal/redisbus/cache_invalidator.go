package redisbus

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type RedisCacheInvalidator struct {
	rdb     *redis.Client
	channel string
}

func NewRedisCacheInvalidator(rdb *redis.Client, channel string) *RedisCacheInvalidator {
	return &RedisCacheInvalidator{rdb: rdb, channel: channel}
}

// publish method
func (r *RedisCacheInvalidator) Publish(ctx context.Context) error {
	err := r.rdb.Publish(ctx, r.channel, "refresh").Err()
	if err != nil {
		slog.Error("Error cache invalidator publishing to Redis:", "error", err)
		return err
	}
	return nil
}

// subscribe method
func (r *RedisCacheInvalidator) Subscribe(ctx context.Context) (<-chan string, error) {
	pubsub := r.rdb.Subscribe(ctx, r.channel)
	ch := make(chan string, 1)

	go func() {
		defer close(ch)
		defer pubsub.Close()

		for {
			select {
			case msg, ok := <-pubsub.Channel():
				if !ok {
					return
				}
				ch <- msg.Payload
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}
