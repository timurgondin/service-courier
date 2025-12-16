package delivery

import (
	"context"
	"log"
	"time"

	orderGateway "service-courier/internal/gateway/order"
)

type OrderWorker struct {
	service *Service
	gateway *orderGateway.Gateway
	clock   Clock
}

func NewOrderWorker(service *Service, gateway *orderGateway.Gateway, clock Clock) *OrderWorker {
	return &OrderWorker{
		service: service,
		gateway: gateway,
		clock:   clock,
	}
}

func (w *OrderWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("[OrderWorker] Starting order polling worker (interval: 5s)")

	for {
		select {
		case <-ctx.Done():
			log.Println("[OrderWorker] Stopping order polling worker...")
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *OrderWorker) process(ctx context.Context) {
	cursor := w.clock.Now().Add(-5 * time.Second)

	orders, err := w.gateway.GetOrders(ctx, cursor)
	if err != nil {
		log.Printf("[OrderWorker] Failed to fetch orders: %v", err)
		return
	}

	if len(orders) == 0 {
		log.Println("[OrderWorker] No new orders")
		return
	}

	latest := cursor
	for _, o := range orders {
		if _, err := w.service.AssignCourier(ctx, o.ID); err != nil {
			log.Printf("[OrderWorker] Failed to assign courier for order %s: %v", o.ID, err)
		} else {
			log.Printf("[OrderWorker] Assigned courier for order %s", o.ID)
		}

		if o.CreatedAt.After(latest) {
			latest = o.CreatedAt
		}
	}

	log.Printf("[OrderWorker] Processed %d orders, cursor updated to %s", len(orders), latest.Format(time.RFC3339))
}
