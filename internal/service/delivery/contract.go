package delivery

import (
	"context"
	"service-courier/internal/model/courier"
	"service-courier/internal/model/delivery"
	"time"
)

type deliveryRepository interface {
	Create(ctx context.Context, deliveryData delivery.Delivery) error
	GetByOrderID(ctx context.Context, orderID string) (*delivery.Delivery, error)
	DeleteByOrderID(ctx context.Context, orderID string) error
	ListExpired(ctx context.Context, now time.Time) ([]delivery.Delivery, error)
	UpdateStatusByIDs(ctx context.Context, ids []int64, status delivery.DeliveryStatus) error
}

type courierRepository interface {
	GetByID(ctx context.Context, id int64) (*courier.Courier, error)
	GetAvailableWithMinDeliveries(ctx context.Context) (*courier.Courier, error)
	Update(ctx context.Context, courierData courier.Courier) error
	UpdateStatusBatch(ctx context.Context, ids []int64, status courier.CourierStatus) error
	IncrementDeliveriesBatch(ctx context.Context, courierIDs []int64) error
}
