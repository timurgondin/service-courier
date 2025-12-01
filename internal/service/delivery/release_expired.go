package delivery

import (
	"context"
	"fmt"
	"log"
	"service-courier/internal/model/courier"
	"service-courier/internal/model/delivery"
)

func (s *Service) ReleaseExpiredCouriers(ctx context.Context) error {
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		expired, err := s.deliveryRepo.ListActiveExpired(ctx, s.clock.Now())
		if err != nil {
			return fmt.Errorf("list expired: %w", err)
		}

		if len(expired) == 0 {
			return nil
		}

		courierIDsMap := make(map[int64]bool)
		deliveryIDs := make([]int64, len(expired))

		for i, d := range expired {
			courierIDsMap[d.CourierID] = true
			deliveryIDs[i] = d.ID
		}

		courierIDs := make([]int64, 0, len(courierIDsMap))
		for id := range courierIDsMap {
			courierIDs = append(courierIDs, id)
		}

		log.Printf("[ReleaseExpiredCouriers] Completing %d deliveries for %d couriers",
			len(deliveryIDs), len(courierIDs))

		if err := s.deliveryRepo.UpdateStatusByIDs(ctx, deliveryIDs, delivery.StatusCompleted); err != nil {
			return fmt.Errorf("update delivery status: %w", err)
		}

		if err := s.courierRepo.UpdateStatusBatch(ctx, courierIDs, courier.StatusAvailable); err != nil {
			return fmt.Errorf("update courier statuses batch: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
