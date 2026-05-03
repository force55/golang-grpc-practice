package alert

import "context"

type CacheInvalidator interface {
	Publish(ctx context.Context) error
	Subscribe(ctx context.Context) (<-chan string, error)
}
