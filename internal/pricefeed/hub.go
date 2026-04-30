package pricefeed

import "context"

type Hub struct {
	subscribers   map[chan PriceUpdate]struct{}
	subscribeCh   chan chan PriceUpdate
	unsubscribeCh chan chan PriceUpdate
	publishCh     chan PriceUpdate
}

func NewHub() *Hub {
	return &Hub{
		subscribers:   make(map[chan PriceUpdate]struct{}),
		subscribeCh:   make(chan chan PriceUpdate),
		unsubscribeCh: make(chan chan PriceUpdate),
		publishCh:     make(chan PriceUpdate),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case ch := <-h.subscribeCh:
			h.subscribers[ch] = struct{}{}

		case ch := <-h.unsubscribeCh:
			delete(h.subscribers, ch)
			close(ch)

		case update := <-h.publishCh:
			for ch := range h.subscribers {
				select {
				case ch <- update:
				default:
					// If the subscriber's channel is full, skip it to avoid blocking
				}
			}

		case <-ctx.Done():
			for ch := range h.subscribers {
				close(ch)
			}
			return
		}
	}
}

func (h *Hub) Subscribe() chan PriceUpdate {
	ch := make(chan PriceUpdate, 2) // Buffered channel to prevent blocking
	h.subscribeCh <- ch
	return ch
}

func (h *Hub) Unsubscribe(ch chan PriceUpdate) {
	h.unsubscribeCh <- ch
}

func (h *Hub) Publish(update PriceUpdate) {
	h.publishCh <- update
}
