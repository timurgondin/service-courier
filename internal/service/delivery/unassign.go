package delivery

import (
	"context"
	"errors"
	"fmt"
	"service-courier/internal/metrics"
	"service-courier/internal/model/courier"
	"service-courier/internal/model/delivery"
)

func (s *Service) UnassignCourier(ctx context.Context, orderID string) (*UnassignResult, error) {
	var result *UnassignResult

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		deliveryData, err := s.deliveryRepo.GetByOrderID(ctx, orderID)
		if err != nil {
			if errors.Is(err, delivery.ErrDeliveryNotFound) {
				return delivery.ErrDeliveryNotFound
			}
			return fmt.Errorf("get delivery: %w", err)
		}

		courierID := deliveryData.CourierID

		if err := s.deliveryRepo.DeleteByOrderID(ctx, orderID); err != nil {
			return fmt.Errorf("delete delivery: %w", err)
		}

		courierData, err := s.courierRepo.GetByID(ctx, courierID)
		if err != nil {
			if errors.Is(err, courier.ErrCourierNotFound) {
				return courier.ErrCourierNotFound
			}
			return fmt.Errorf("get courier: %w", err)
		}

		courierData.Status = courier.StatusAvailable
		if err := s.courierRepo.Update(ctx, *courierData); err != nil {
			if errors.Is(err, courier.ErrCourierNotFound) {
				return courier.ErrCourierNotFound
			}
			return fmt.Errorf("update courier status: %w", err)
		}

		result = &UnassignResult{
			OrderID:   orderID,
			Status:    delivery.StatusUnassigned,
			CourierID: courierID,
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unassign courier transaction: %w", err)
	}

	metrics.OpsCounter.Inc()

	return result, nil
}
