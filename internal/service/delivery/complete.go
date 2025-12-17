package delivery

import (
	"context"
	"errors"
	"fmt"
	"service-courier/internal/model/courier"
	modelDelivery "service-courier/internal/model/delivery"
)

func (s *Service) CompleteDelivery(ctx context.Context, orderID string) error {
	return s.txManager.Do(ctx, func(ctx context.Context) error {
		deliveryData, err := s.deliveryRepo.GetByOrderID(ctx, orderID)
		if err != nil {
			if errors.Is(err, modelDelivery.ErrDeliveryNotFound) {
				return modelDelivery.ErrDeliveryNotFound
			}
			return fmt.Errorf("get delivery: %w", err)
		}

		if err := s.deliveryRepo.UpdateStatusByIDs(ctx, []int64{deliveryData.ID}, modelDelivery.StatusCompleted); err != nil {
			return fmt.Errorf("update delivery status: %w", err)
		}

		if err := s.courierRepo.UpdateStatusBatch(ctx, []int64{deliveryData.CourierID}, courier.StatusAvailable); err != nil {
			if errors.Is(err, courier.ErrCourierNotFound) {
				return courier.ErrCourierNotFound
			}
			return fmt.Errorf("update courier status: %w", err)
		}

		return nil
	})
}
