package delivery

import (
	"context"
	"service-courier/internal/model/courier"
	"service-courier/internal/model/delivery"
)

type deliveryRepository interface {
	Create(ctx context.Context, deliveryData delivery.Delivery) error
	GetByOrderID(ctx context.Context, orderID string) (*delivery.Delivery, error)
	DeleteByOrderID(ctx context.Context, orderID string) error
}

type courierRepository interface {
	GetByID(ctx context.Context, id int64) (*courier.Courier, error)
	GetAvailable(ctx context.Context) (*courier.Courier, error)
	Update(ctx context.Context, courierData courier.Courier) error
}
