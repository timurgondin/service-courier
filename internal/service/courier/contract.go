package courier

import (
	"context"
	"service-courier/internal/model/courier"
)

type courierRepository interface {
	GetByID(ctx context.Context, id int64) (*courier.Courier, error)
	GetAll(ctx context.Context) ([]courier.Courier, error)
	Create(ctx context.Context, courierData courier.Courier) (int64, error)
	Update(ctx context.Context, courierData courier.Courier) error
	GetAvailable(ctx context.Context) (*courier.Courier, error)
}
