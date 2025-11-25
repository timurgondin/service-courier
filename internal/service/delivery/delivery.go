package delivery

import (
	"context"
	"errors"
	"fmt"
	"log"
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
		_, err := s.deliveryRepo.GetByOrderID(ctx, orderID)
		if err == nil {
			return delivery.ErrOrderAlreadyAssigned
		}
		if !errors.Is(err, delivery.ErrDeliveryNotFound) {
			return fmt.Errorf("check existing delivery: %w", err)
		}

		availableCourier, err := s.courierRepo.GetAvailableWithMinDeliveries(ctx)
		if err != nil {
			return courier.ErrNoAvailableCouriers
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
			TransportType:    availableCourier.TransportType,
			DeliveryDeadline: deadline,
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("assign courier transaction: %w", err)
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

		courierID := deliveryData.CourierID

		if err := s.deliveryRepo.DeleteByOrderID(ctx, orderID); err != nil {
			return fmt.Errorf("delete delivery: %w", err)
		}

		courierData, err := s.courierRepo.GetByID(ctx, courierID)
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
			CourierID: courierID,
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("unassign courier transaction: %w", err)
	}

	return result, nil
}

func (s *Service) ReleaseExpiredCouriers(ctx context.Context) error {
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		expired, err := s.deliveryRepo.ListExpired(ctx, time.Now())
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

		if err := s.courierRepo.IncrementDeliveriesBatch(ctx, courierIDs); err != nil {
			return fmt.Errorf("increment deliveries batch: %w", err)
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
