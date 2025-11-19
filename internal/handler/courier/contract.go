package courier

import (
	"context"
	"service-courier/internal/model/courier"
)

type courierService interface {
	GetCourier(ctx context.Context, id int64) (*courier.Courier, error)
	GetAllCouriers(ctx context.Context) ([]courier.Courier, error)
	CreateCourier(ctx context.Context, courier courier.Courier) (id int64, err error)
	UpdateCourier(ctx context.Context, courier courier.Courier) error
}
