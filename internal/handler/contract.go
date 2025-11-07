package handler

import (
	"context"
	"service-courier/internal/model"
)

type CourierService interface {
	GetCourier(ctx context.Context, id int64) (*model.Courier, error)
	GetAllCouriers(ctx context.Context) ([]model.Courier, error)
	CreateCourier(ctx context.Context, req *model.CourierCreateRequest) (int64, error)
	UpdateCourier(ctx context.Context, req *model.CourierUpdateRequest) error
}
