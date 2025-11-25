package courier

import (
	"context"
	"service-courier/internal/model/courier"
)

type Service struct {
	repo courierRepository
}

func NewCourierService(repo courierRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCourier(ctx context.Context, id int64) (*courier.Courier, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetAllCouriers(ctx context.Context) ([]courier.Courier, error) {
	return s.repo.GetAll(ctx)
}

func (s *Service) CreateCourier(ctx context.Context, courierData courier.Courier) (id int64, err error) {
	return s.repo.Create(ctx, courierData)
}

func (s *Service) UpdateCourier(ctx context.Context, courierData courier.Courier) error {
	return s.repo.Update(ctx, courierData)
}
