package courier

import (
	"context"
	"service-courier/internal/model"
)

type courierRepository interface {
	GetByID(ctx context.Context, id int64) (*model.CourierDB, error)
	GetAll(ctx context.Context) ([]model.CourierDB, error)
	Create(ctx context.Context, courier *model.CourierDB) (int64, error)
	Update(ctx context.Context, courier *model.CourierUpdateDB) error
}
