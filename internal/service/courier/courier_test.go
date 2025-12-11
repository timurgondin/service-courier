package courier_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	model "service-courier/internal/model/courier"
	courierService "service-courier/internal/service/courier"
	"service-courier/internal/service/courier/mocks"
)

func TestGetCourier_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	expectedCourier := &model.Courier{
		ID:            1,
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusAvailable,
		TransportType: model.TransportCar,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	mockRepo.EXPECT().
		GetByID(gomock.Any(), int64(1)).
		Return(expectedCourier, nil)

	result, err := service.GetCourier(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.ID != 1 {
		t.Fatalf("expected ID=1, got %d", result.ID)
	}
	if result.Name != "Ivan" {
		t.Fatalf("expected Name=Ivan, got %s", result.Name)
	}
	if result.Phone != "+78005553535" {
		t.Fatalf("expected Phone=+78005553535, got %s", result.Phone)
	}
	if result.Status != model.StatusAvailable {
		t.Fatalf("expected Status=available, got %s", result.Status)
	}
	if result.TransportType != model.TransportCar {
		t.Fatalf("expected TransportType=car, got %s", result.TransportType)
	}
}

func TestGetCourier_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	mockRepo.EXPECT().
		GetByID(gomock.Any(), int64(999)).
		Return(nil, model.ErrCourierNotFound)

	result, err := service.GetCourier(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, model.ErrCourierNotFound) {
		t.Fatalf("expected ErrCourierNotFound, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestGetCourier_RepositoryError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	repoErr := errors.New("database connection error")
	mockRepo.EXPECT().
		GetByID(gomock.Any(), int64(1)).
		Return(nil, repoErr)

	result, err := service.GetCourier(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected database error, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestGetAllCouriers_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	expectedCouriers := []model.Courier{
		{
			ID:            1,
			Name:          "Ivan",
			Phone:         "+78005553535",
			Status:        model.StatusAvailable,
			TransportType: model.TransportCar,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            2,
			Name:          "Petr",
			Phone:         "+78005553536",
			Status:        model.StatusBusy,
			TransportType: model.TransportScooter,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	mockRepo.EXPECT().
		GetAll(gomock.Any()).
		Return(expectedCouriers, nil)

	result, err := service.GetAllCouriers(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 couriers, got %d", len(result))
	}
	if result[0].ID != 1 {
		t.Fatalf("expected first courier ID=1, got %d", result[0].ID)
	}
	if result[1].ID != 2 {
		t.Fatalf("expected second courier ID=2, got %d", result[1].ID)
	}
}

func TestGetAllCouriers_EmptyList(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	mockRepo.EXPECT().
		GetAll(gomock.Any()).
		Return([]model.Courier{}, nil)

	result, err := service.GetAllCouriers(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 couriers, got %d", len(result))
	}
}

func TestGetAllCouriers_RepositoryError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	repoErr := errors.New("database query error")
	mockRepo.EXPECT().
		GetAll(gomock.Any()).
		Return(nil, repoErr)

	result, err := service.GetAllCouriers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected database error, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestCreateCourier_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	courierData := model.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusAvailable,
		TransportType: model.TransportCar,
	}

	expectedID := int64(10)
	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, c model.Courier) (int64, error) {
			if c.Name != courierData.Name {
				t.Errorf("expected Name=%s, got %s", courierData.Name, c.Name)
			}
			if c.Phone != courierData.Phone {
				t.Errorf("expected Phone=%s, got %s", courierData.Phone, c.Phone)
			}
			if c.Status != courierData.Status {
				t.Errorf("expected Status=%s, got %s", courierData.Status, c.Status)
			}
			if c.TransportType != courierData.TransportType {
				t.Errorf("expected TransportType=%s, got %s", courierData.TransportType, c.TransportType)
			}
			return expectedID, nil
		})

	id, err := service.CreateCourier(context.Background(), courierData)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != expectedID {
		t.Fatalf("expected ID=%d, got %d", expectedID, id)
	}
}

func TestCreateCourier_PhoneExists(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	courierData := model.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusAvailable,
		TransportType: model.TransportCar,
	}

	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(int64(0), model.ErrPhoneExists)

	id, err := service.CreateCourier(context.Background(), courierData)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, model.ErrPhoneExists) {
		t.Fatalf("expected ErrPhoneExists, got %v", err)
	}
	if id != 0 {
		t.Fatalf("expected ID=0, got %d", id)
	}
}

func TestCreateCourier_RepositoryError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	courierData := model.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusAvailable,
		TransportType: model.TransportCar,
	}

	repoErr := errors.New("database insert error")
	mockRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(int64(0), repoErr)

	id, err := service.CreateCourier(context.Background(), courierData)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected database error, got %v", err)
	}
	if id != 0 {
		t.Fatalf("expected ID=0, got %d", id)
	}
}

func TestUpdateCourier_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	courierData := model.Courier{
		ID:     1,
		Name:   "Updated Name",
		Phone:  "+78005553535",
		Status: model.StatusBusy,
	}

	mockRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, c model.Courier) error {
			if c.ID != courierData.ID {
				t.Errorf("expected ID=%d, got %d", courierData.ID, c.ID)
			}
			if c.Name != courierData.Name {
				t.Errorf("expected Name=%s, got %s", courierData.Name, c.Name)
			}
			return nil
		})

	err := service.UpdateCourier(context.Background(), courierData)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateCourier_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	courierData := model.Courier{
		ID:    999,
		Name:  "Updated Name",
		Phone: "+78005553535",
	}

	mockRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(model.ErrCourierNotFound)

	err := service.UpdateCourier(context.Background(), courierData)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, model.ErrCourierNotFound) {
		t.Fatalf("expected ErrCourierNotFound, got %v", err)
	}
}

func TestUpdateCourier_PhoneExists(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	courierData := model.Courier{
		ID:    1,
		Name:  "Updated Name",
		Phone: "+78005553536",
	}

	mockRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(model.ErrPhoneExists)

	err := service.UpdateCourier(context.Background(), courierData)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, model.ErrPhoneExists) {
		t.Fatalf("expected ErrPhoneExists, got %v", err)
	}
}

func TestUpdateCourier_RepositoryError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockcourierRepository(ctrl)
	service := courierService.NewCourierService(mockRepo)

	courierData := model.Courier{
		ID:    1,
		Name:  "Updated Name",
		Phone: "+78005553535",
	}

	repoErr := errors.New("database update error")
	mockRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(repoErr)

	err := service.UpdateCourier(context.Background(), courierData)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected database error, got %v", err)
	}
}
