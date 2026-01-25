package changed

import (
	"context"
	"errors"
	"fmt"
	modelDelivery "service-courier/internal/model/delivery"
	"service-courier/internal/model/order"
)

type Usecase struct {
	delivery deliveryService
}

func NewUsecase(delivery deliveryService) *Usecase {
	return &Usecase{
		delivery: delivery,
	}
}

func (u *Usecase) Process(ctx context.Context, o order.Order) error {
	switch o.Status {
	case order.StatusCreated:
		if _, err := u.delivery.AssignCourier(ctx, o.ID); err != nil && !errors.Is(err, modelDelivery.ErrOrderAlreadyAssigned) {
			return fmt.Errorf("assign courier: %w", err)
		}
	case order.StatusCancelled:
		if _, err := u.delivery.UnassignCourier(ctx, o.ID); err != nil && !errors.Is(err, modelDelivery.ErrDeliveryNotFound) {
			return fmt.Errorf("unassign courier: %w", err)
		}
	case order.StatusCompleted:
		if err := u.delivery.CompleteDelivery(ctx, o.ID); err != nil && !errors.Is(err, modelDelivery.ErrDeliveryNotFound) {
			return fmt.Errorf("complete delivery: %w", err)
		}
	default:
		return nil
	}

	return nil
}
