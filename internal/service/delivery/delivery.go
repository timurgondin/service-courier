package delivery

import (
	"context"
	"errors"
	"fmt"
	"service-courier/internal/model/courier"
	"service-courier/internal/model/delivery"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

type Service struct {
	deliveryRepo deliveryRepository
	courierRepo  courierRepository
	timeFactory  *DeliveryTimeFactory
	txManager    *manager.Manager
}

func NewDeliveryService(
	deliveryRepo deliveryRepository,
	courierRepo courierRepository,
	timeFactory *DeliveryTimeFactory,
	txManager *manager.Manager,
) *Service {
	return &Service{
		deliveryRepo: deliveryRepo,
		courierRepo:  courierRepo,
		timeFactory:  timeFactory,
		txManager:    txManager,
	}

}

func (s *Service) AssignCourier(ctx context.Context, orderID string) (*AssignResult, error) {
	var result *AssignResult

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		existingDelivery, err := s.deliveryRepo.GetByOrderID(ctx, orderID)
		if err == nil && existingDelivery != nil {
			return delivery.ErrOrderAlreadyAssigned
		}
		availableCourier, err := s.courierRepo.GetAvailable(ctx)
		if err != nil {
			return delivery.ErrNoAvailableCouriers
		}
		assignedAt := time.Now()
		deadline := s.timeFactory.CalculateDeadline(availableCourier.TransportType, assignedAt)
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
			return fmt.Errorf("update courier status: %w", err)
		}
		result = &AssignResult{
			CourierID:        availableCourier.ID,
			OrderID:          orderID,
			TransportType:    courier.TransportType(availableCourier.TransportType),
			DeliveryDeadline: deadline,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil

}

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
		if err := s.deliveryRepo.DeleteByOrderID(ctx, orderID); err != nil {
			return fmt.Errorf("delete delivery: %w", err)
		}
		courierData, err := s.courierRepo.GetByID(ctx, deliveryData.CourierID)
		if err != nil {
			return fmt.Errorf("get courier: %w", err)
		}

		courierData.Status = courier.StatusAvailable
		if err := s.courierRepo.Update(ctx, *courierData); err != nil {
			return fmt.Errorf("update courier status: %w", err)
		}

		result = &UnassignResult{
			OrderID:   orderID,
			Status:    "unassigned",
			CourierID: deliveryData.CourierID,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
