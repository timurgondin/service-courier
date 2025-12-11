//go:generate mockgen -source=contract.go -destination=./mocks/delivery_service_mock.go -package=mocks

package delivery

import (
	"context"
	"service-courier/internal/service/delivery"
)

type deliveryService interface {
	AssignCourier(ctx context.Context, orderID string) (*delivery.AssignResult, error)
	UnassignCourier(ctx context.Context, orderID string) (*delivery.UnassignResult, error)
}
