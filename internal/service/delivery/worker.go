package delivery

import (
	"context"
	"log"
	"time"
)

type Worker struct {
	service  *Service
	interval time.Duration
}

func NewWorker(service *Service, interval time.Duration) *Worker {
	return &Worker{
		service:  service,
		interval: interval,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("[Worker] Starting courier release worker (interval: %v)", w.interval)

	if err := w.service.ReleaseExpiredCouriers(ctx); err != nil {
		log.Printf("[Worker] Failed to release expired couriers on startup: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("[Worker] Stopping courier release worker...")
			return
		case <-ticker.C:
			if err := w.service.ReleaseExpiredCouriers(ctx); err != nil {
				log.Printf("[Worker] Failed to release expired couriers: %v", err)
			}
		}
	}
}
