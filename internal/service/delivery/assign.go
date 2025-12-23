package delivery

import (
	"context"
	"errors"
	"fmt"
	"service-courier/internal/metrics"
	"service-courier/internal/model/courier"
	"service-courier/internal/model/delivery"
)

func (s *Service) AssignCourier(ctx context.Context, orderID string) (*AssignResult, error) {
	var result *AssignResult

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		existingDelivery, err := s.deliveryRepo.GetByOrderID(ctx, orderID)
		if err != nil && !errors.Is(err, delivery.ErrDeliveryNotFound) {
			return fmt.Errorf("check existing delivery: %w", err)
		}

		if existingDelivery != nil {
			return delivery.ErrOrderAlreadyAssigned
		}

		availableCourier, err := s.courierRepo.GetAvailableWithMinDeliveries(ctx)
		if err != nil {
			if errors.Is(err, courier.ErrNoAvailableCouriers) {
				return courier.ErrNoAvailableCouriers
			}
			return fmt.Errorf("get available courier: %w", err)
		}

		assignedAt := s.clock.Now()

		transport := s.transportFactory.Create(availableCourier.TransportType)

		deadline := assignedAt.Add(transport.DeliveryDuration())

		deliveryData := delivery.Delivery{
			CourierID:  availableCourier.ID,
			OrderID:    orderID,
			AssignedAt: assignedAt,
			Deadline:   deadline,
		}

		if err := s.deliveryRepo.Create(ctx, deliveryData); err != nil {
			return fmt.Errorf("create delivery: %w", err)
		}

		availableCourier.Status = courier.StatusBusy
		if err := s.courierRepo.Update(ctx, *availableCourier); err != nil {
			if errors.Is(err, courier.ErrCourierNotFound) {
				return courier.ErrCourierNotFound
			}
			return fmt.Errorf("update courier status: %w", err)
		}

		result = &AssignResult{
			CourierID:     availableCourier.ID,
			OrderID:       orderID,
			TransportType: availableCourier.TransportType,
			Deadline:      deadline,
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("assign courier transaction: %w", err)
	}

	metrics.OpsCounter.Inc()
	return result, nil
}
