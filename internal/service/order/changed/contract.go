package changed

import (
	"context"
	"service-courier/internal/service/delivery"
)

type deliveryService interface {
	AssignCourier(ctx context.Context, orderID string) (*delivery.AssignResult, error)
	UnassignCourier(ctx context.Context, orderID string) (*delivery.UnassignResult, error)
	CompleteDelivery(ctx context.Context, orderID string) error
}
