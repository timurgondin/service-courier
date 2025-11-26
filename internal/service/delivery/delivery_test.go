package delivery_test

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"service-courier/internal/mocks"
	modelCourier "service-courier/internal/model/courier"
	modelDelivery "service-courier/internal/model/delivery"
	deliveryService "service-courier/internal/service/delivery"
)

func TestAssignCourier_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	orderID := "order-1"
	availableCourier := &modelCourier.Courier{
		ID:            10,
		Name:          "Ivan",
		Status:        modelCourier.StatusAvailable,
		TransportType: modelCourier.TransportCar,
	}

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		GetByOrderID(gomock.Any(), orderID).
		Return(nil, modelDelivery.ErrDeliveryNotFound)

	mockCourierRepo.EXPECT().
		GetAvailableWithMinDeliveries(gomock.Any()).
		Return(availableCourier, nil)

	mockDeliveryRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil)

	mockCourierRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(nil)

	result, err := service.AssignCourier(context.Background(), orderID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.OrderID != orderID {
		t.Fatalf("expected OrderID=%s, got %s", orderID, result.OrderID)
	}
	if result.CourierID != 10 {
		t.Fatalf("expected CourierID=10, got %d", result.CourierID)
	}
	if result.TransportType != modelCourier.TransportCar {
		t.Fatalf("expected TransportType=car, got %s", result.TransportType)
	}
	if result.Deadline.IsZero() {
		t.Fatal("expected deadline to be set, got zero time")
	}
}

func TestAssignCourier_OrderAlreadyAssigned(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	orderID := "order-1"
	existingDelivery := &modelDelivery.Delivery{
		ID:        1,
		OrderID:   orderID,
		CourierID: 5,
		Status:    modelDelivery.StatusActive,
	}

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		GetByOrderID(gomock.Any(), orderID).
		Return(existingDelivery, nil)

	result, err := service.AssignCourier(context.Background(), orderID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, modelDelivery.ErrOrderAlreadyAssigned) {
		t.Fatalf("expected ErrOrderAlreadyAssigned, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestAssignCourier_NoAvailableCouriers(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	orderID := "order-1"

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		GetByOrderID(gomock.Any(), orderID).
		Return(nil, modelDelivery.ErrDeliveryNotFound)

	mockCourierRepo.EXPECT().
		GetAvailableWithMinDeliveries(gomock.Any()).
		Return(nil, modelCourier.ErrNoAvailableCouriers)

	result, err := service.AssignCourier(context.Background(), orderID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, modelCourier.ErrNoAvailableCouriers) {
		t.Fatalf("expected ErrNoAvailableCouriers, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestAssignCourier_GetByOrderIDError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	orderID := "order-1"
	repoErr := errors.New("database error")

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		GetByOrderID(gomock.Any(), orderID).
		Return(nil, repoErr)

	result, err := service.AssignCourier(context.Background(), orderID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestAssignCourier_CreateDeliveryError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	orderID := "order-1"
	availableCourier := &modelCourier.Courier{
		ID:            10,
		Status:        modelCourier.StatusAvailable,
		TransportType: modelCourier.TransportCar,
	}
	repoErr := errors.New("create delivery error")

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		GetByOrderID(gomock.Any(), orderID).
		Return(nil, modelDelivery.ErrDeliveryNotFound)

	mockCourierRepo.EXPECT().
		GetAvailableWithMinDeliveries(gomock.Any()).
		Return(availableCourier, nil)

	mockDeliveryRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(repoErr)

	result, err := service.AssignCourier(context.Background(), orderID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestAssignCourier_UpdateCourierError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	orderID := "order-1"
	availableCourier := &modelCourier.Courier{
		ID:            10,
		Status:        modelCourier.StatusAvailable,
		TransportType: modelCourier.TransportCar,
	}
	repoErr := errors.New("update courier error")

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		GetByOrderID(gomock.Any(), orderID).
		Return(nil, modelDelivery.ErrDeliveryNotFound)

	mockCourierRepo.EXPECT().
		GetAvailableWithMinDeliveries(gomock.Any()).
		Return(availableCourier, nil)

	mockDeliveryRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil)

	mockCourierRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(repoErr)

	result, err := service.AssignCourier(context.Background(), orderID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestUnassignCourier_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	orderID := "order-1"
	deliveryData := &modelDelivery.Delivery{
		ID:        1,
		OrderID:   orderID,
		CourierID: 10,
		Status:    modelDelivery.StatusActive,
	}

	courierData := &modelCourier.Courier{
		ID:     10,
		Name:   "Ivan",
		Status: modelCourier.StatusBusy,
	}

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		GetByOrderID(gomock.Any(), orderID).
		Return(deliveryData, nil)

	mockDeliveryRepo.EXPECT().
		DeleteByOrderID(gomock.Any(), orderID).
		Return(nil)

	mockCourierRepo.EXPECT().
		GetByID(gomock.Any(), int64(10)).
		Return(courierData, nil)

	mockCourierRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(nil)

	result, err := service.UnassignCourier(context.Background(), orderID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.OrderID != orderID {
		t.Fatalf("expected OrderID=%s, got %s", orderID, result.OrderID)
	}
	if result.CourierID != 10 {
		t.Fatalf("expected CourierID=10, got %d", result.CourierID)
	}
	if result.Status != "unassigned" {
		t.Fatalf("expected Status=unassigned, got %s", result.Status)
	}
}

func TestUnassignCourier_DeliveryNotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	orderID := "order-1"

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		GetByOrderID(gomock.Any(), orderID).
		Return(nil, modelDelivery.ErrDeliveryNotFound)

	result, err := service.UnassignCourier(context.Background(), orderID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, modelDelivery.ErrDeliveryNotFound) {
		t.Fatalf("expected ErrDeliveryNotFound, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestUnassignCourier_GetByOrderIDError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	orderID := "order-1"
	repoErr := errors.New("database error")

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		GetByOrderID(gomock.Any(), orderID).
		Return(nil, repoErr)

	result, err := service.UnassignCourier(context.Background(), orderID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestUnassignCourier_DeleteError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	orderID := "order-1"
	deliveryData := &modelDelivery.Delivery{
		ID:        1,
		OrderID:   orderID,
		CourierID: 10,
		Status:    modelDelivery.StatusActive,
	}
	repoErr := errors.New("delete error")

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		GetByOrderID(gomock.Any(), orderID).
		Return(deliveryData, nil)

	mockDeliveryRepo.EXPECT().
		DeleteByOrderID(gomock.Any(), orderID).
		Return(repoErr)

	result, err := service.UnassignCourier(context.Background(), orderID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestUnassignCourier_GetCourierError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	orderID := "order-1"
	deliveryData := &modelDelivery.Delivery{
		ID:        1,
		OrderID:   orderID,
		CourierID: 10,
		Status:    modelDelivery.StatusActive,
	}
	repoErr := errors.New("get courier error")

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		GetByOrderID(gomock.Any(), orderID).
		Return(deliveryData, nil)

	mockDeliveryRepo.EXPECT().
		DeleteByOrderID(gomock.Any(), orderID).
		Return(nil)

	mockCourierRepo.EXPECT().
		GetByID(gomock.Any(), int64(10)).
		Return(nil, repoErr)

	result, err := service.UnassignCourier(context.Background(), orderID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestUnassignCourier_UpdateCourierError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	orderID := "order-1"
	deliveryData := &modelDelivery.Delivery{
		ID:        1,
		OrderID:   orderID,
		CourierID: 10,
		Status:    modelDelivery.StatusActive,
	}
	courierData := &modelCourier.Courier{
		ID:     10,
		Status: modelCourier.StatusBusy,
	}
	repoErr := errors.New("update courier error")

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		GetByOrderID(gomock.Any(), orderID).
		Return(deliveryData, nil)

	mockDeliveryRepo.EXPECT().
		DeleteByOrderID(gomock.Any(), orderID).
		Return(nil)

	mockCourierRepo.EXPECT().
		GetByID(gomock.Any(), int64(10)).
		Return(courierData, nil)

	mockCourierRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(repoErr)

	result, err := service.UnassignCourier(context.Background(), orderID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestReleaseExpiredCouriers_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	expiredDeliveries := []modelDelivery.Delivery{
		{ID: 1, CourierID: 10, OrderID: "order-1", Status: modelDelivery.StatusActive},
		{ID: 2, CourierID: 10, OrderID: "order-2", Status: modelDelivery.StatusActive},
		{ID: 3, CourierID: 20, OrderID: "order-3", Status: modelDelivery.StatusActive},
	}

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		ListActiveExpired(gomock.Any(), gomock.Any()).
		Return(expiredDeliveries, nil)

	mockDeliveryRepo.EXPECT().
		UpdateStatusByIDs(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	mockCourierRepo.EXPECT().
		UpdateStatusBatch(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	err := service.ReleaseExpiredCouriers(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestReleaseExpiredCouriers_NoExpired(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		ListActiveExpired(gomock.Any(), gomock.Any()).
		Return([]modelDelivery.Delivery{}, nil)

	err := service.ReleaseExpiredCouriers(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestReleaseExpiredCouriers_ListError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	repoErr := errors.New("list expired error")

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		ListActiveExpired(gomock.Any(), gomock.Any()).
		Return(nil, repoErr)

	err := service.ReleaseExpiredCouriers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReleaseExpiredCouriers_UpdateDeliveryStatusError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	expiredDeliveries := []modelDelivery.Delivery{
		{ID: 1, CourierID: 10, OrderID: "order-1", Status: modelDelivery.StatusActive},
	}
	repoErr := errors.New("update delivery status error")

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		ListActiveExpired(gomock.Any(), gomock.Any()).
		Return(expiredDeliveries, nil)

	mockDeliveryRepo.EXPECT().
		UpdateStatusByIDs(gomock.Any(), []int64{1}, gomock.Any()).
		Return(repoErr)

	err := service.ReleaseExpiredCouriers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReleaseExpiredCouriers_UpdateCourierStatusError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeliveryRepo := mocks.NewMockdeliveryRepository(ctrl)
	mockCourierRepo := mocks.NewMockDeliveryCourierRepository(ctrl)
	mockTxManager := mocks.NewMocktransactionManager(ctrl)
	timeFactory := deliveryService.NewDeliveryTimeFactory()

	service := deliveryService.NewDeliveryService(
		mockDeliveryRepo,
		mockCourierRepo,
		timeFactory,
		mockTxManager,
	)

	expiredDeliveries := []modelDelivery.Delivery{
		{ID: 1, CourierID: 10, OrderID: "order-1", Status: modelDelivery.StatusActive},
	}
	repoErr := errors.New("update courier status error")

	mockTxManager.EXPECT().
		Do(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	mockDeliveryRepo.EXPECT().
		ListActiveExpired(gomock.Any(), gomock.Any()).
		Return(expiredDeliveries, nil)

	mockDeliveryRepo.EXPECT().
		UpdateStatusByIDs(gomock.Any(), []int64{1}, gomock.Any()).
		Return(nil)

	mockCourierRepo.EXPECT().
		UpdateStatusBatch(gomock.Any(), []int64{10}, gomock.Any()).
		Return(repoErr)

	err := service.ReleaseExpiredCouriers(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
