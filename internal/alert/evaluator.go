package alert

import (
	"context"
	"crypto-notifier/internal/pricefeed"
	"log/slog"
	"strconv"
)

type Evaluator struct {
	hub         *pricefeed.Hub
	queries     *Queries
	NotifCh     chan TriggeredInfo
	cache       map[string][]Alert
	invalidator CacheInvalidator
}

type TriggeredInfo struct {
	AlertID   string
	Symbol    string
	Price     float64
	Direction string
}

func NewEvaluator(hub *pricefeed.Hub, queries *Queries, invalidator CacheInvalidator) *Evaluator {
	return &Evaluator{
		hub:         hub,
		queries:     queries,
		NotifCh:     make(chan TriggeredInfo, 16),
		cache:       make(map[string][]Alert),
		invalidator: invalidator,
	}
}

func (e *Evaluator) Run(ctx context.Context) {
	ch := e.hub.Subscribe()
	defer e.hub.Unsubscribe(ch)

	err := e.LoadCache(ctx)
	if err != nil {
		slog.Error("Error loading alerts:", "error", err)
		return
	}

	invalidateCh, _ := e.invalidator.Subscribe(ctx)

	for {
		select {
		case update := <-ch:
			alerts := e.cache[update.Symbol]

			for _, alert := range alerts {
				price, err := strconv.ParseFloat(alert.TargetPrice, 64)
				if err != nil {
					slog.Error("Error parsing price for alert:", "error", err, "alert_id", alert.ID)
					continue
				}
				if (alert.Direction == "below" && update.Price <= price) || (alert.Direction == "above" && update.Price >= price) {
					_, err := e.queries.CreateTriggeredAlert(ctx, CreateTriggeredAlertParams{
						AlertID:        alert.ID,
						TriggeredPrice: alert.TargetPrice,
					})
					if err != nil {
						slog.Error("Error creating triggered alert:", "error", err)
					}

					err = e.queries.DeactivateAlert(ctx, alert.ID)
					if err != nil {
						slog.Error("Error deactivating alert:", "error", err)
					}

					e.NotifCh <- TriggeredInfo{
						AlertID:   alert.ID.String(),
						Symbol:    alert.Symbol,
						Price:     update.Price,
						Direction: alert.Direction,
					}

					_ = e.LoadCache(ctx)
					break
				}
			}
		case <-invalidateCh:
			if err := e.LoadCache(ctx); err != nil {
				slog.Error("Error refreshing alerts:", "error", err)
			}
		// Handle triggered notifications (e.g., log, send to user)
		case <-ctx.Done():
			return
		}
	}
}

func (e *Evaluator) LoadCache(ctx context.Context) error {
	alerts, err := e.queries.GetAllActiveAlerts(ctx)
	if err != nil {
		return err
	}

	newCache := make(map[string][]Alert)
	for _, alert := range alerts {
		newCache[alert.Symbol] = append(newCache[alert.Symbol], alert)
	}

	e.cache = newCache

	return nil
}
