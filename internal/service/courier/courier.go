package courier

import (
	"context"
	"errors"
	"service-courier/internal/model"
	"slices"
	"time"
	"unicode"
)

type CourierService struct {
	repository courierRepository
}

func NewCourierService(repository courierRepository) *CourierService {
	return &CourierService{repository: repository}
}

func (c *CourierService) GetCourier(ctx context.Context, id int64) (*model.Courier, error) {
	courierDB, err := c.repository.GetByID(ctx, id)

	if err != nil {
		if errors.Is(err, model.ErrCourierNotFound) {
			return nil, model.ErrCourierNotFound
		}
		return nil, err
	}

	courier := &model.Courier{
		ID:     courierDB.ID,
		Name:   courierDB.Name,
		Phone:  courierDB.Phone,
		Status: courierDB.Status,
	}

	return courier, nil
}

func (c *CourierService) GetAllCouriers(ctx context.Context) ([]model.Courier, error) {
	couriersDB, err := c.repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var couriers []model.Courier
	for _, courierDB := range couriersDB {
		courier := model.Courier{
			ID:     courierDB.ID,
			Name:   courierDB.Name,
			Phone:  courierDB.Phone,
			Status: courierDB.Status,
		}
		couriers = append(couriers, courier)
	}

	if couriers == nil {
		couriers = []model.Courier{}
	}

	return couriers, nil
}

func (c *CourierService) CreateCourier(ctx context.Context, req *model.CourierCreateRequest) (id int64, err error) {

	if err = validateName(req.Name); err != nil {
		return
	}
	if err = validatePhone(req.Phone); err != nil {
		return
	}
	if err = validateStatus(req.Status); err != nil {
		return
	}

	courierDB := &model.CourierDB{
		Name:   req.Name,
		Phone:  req.Phone,
		Status: req.Status,
	}

	id, err = c.repository.Create(ctx, courierDB)

	if err != nil {
		return
	}

	return
}

func (u *CourierService) UpdateCourier(ctx context.Context, req *model.CourierUpdateRequest) error {

	if req.Name != nil {
		if err := validateName(*req.Name); err != nil {
			return err
		}
	}
	if req.Phone != nil {
		if err := validatePhone(*req.Phone); err != nil {
			return err
		}
	}
	if req.Status != nil {
		if err := validateStatus(*req.Status); err != nil {
			return err
		}
	}

	courierUpdateDB := &model.CourierUpdateDB{
		ID:        req.ID,
		Name:      req.Name,
		Phone:     req.Phone,
		Status:    req.Status,
		UpdatedAt: time.Now(),
	}

	err := u.repository.Update(ctx, courierUpdateDB)

	if err != nil {
		switch {
		case errors.Is(err, model.ErrCourierNotFound):
			return model.ErrCourierNotFound
		case errors.Is(err, model.ErrPhoneExists):
			return model.ErrPhoneExists
		}
		return err
	}

	return nil
}

func validateName(name string) error {
	if name == "" {
		return model.ErrCourierNameEmpty
	}
	return nil
}

func validatePhone(phone string) error {
	if phone == "" {
		return model.ErrCourierPhoneEmpty
	}

	if len(phone) != 12 {
		return model.ErrCourierPhoneInvalid
	}

	if phone[0] != '+' {
		return model.ErrCourierPhoneInvalid
	}

	digits := phone[1:]
	for _, digit := range digits {
		if !unicode.IsDigit(digit) {
			return model.ErrCourierPhoneInvalid
		}
	}

	return nil
}

func validateStatus(status string) error {
	if status == "" {
		return model.ErrCourierStatusEmpty
	}
	if !slices.Contains(model.ValidCourierStatuses, status) {
		return model.ErrCourierStatusInvalid
	}
	return nil
}
