package delivery_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"service-courier/internal/integration"
	modelCourier "service-courier/internal/model/courier"
	modelDelivery "service-courier/internal/model/delivery"
	courierRepo "service-courier/internal/repository/courier"
	deliveryRepo "service-courier/internal/repository/delivery"
	deliveryService "service-courier/internal/service/delivery"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
)

func TestDeliveryService_AssignCourier_Integration(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	deliveryRepository := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	courierRepository := courierRepo.NewCourierRepository(pool)
	transportFactory := deliveryService.NewTransportFactory()
	txManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	fixed := time.Date(2024, 1, 1, 12, 00, 00, 0, time.UTC)
	clock := deliveryService.NewFixedClock(fixed)

	service := deliveryService.NewDeliveryService(
		deliveryRepository,
		courierRepository,
		transportFactory,
		txManager,
		clock,
	)
	ctx := context.Background()

	// Создаем доступного курьера
	courierData := modelCourier.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        modelCourier.StatusAvailable,
		TransportType: modelCourier.TransportCar,
	}
	courierID, err := courierRepository.Create(ctx, courierData)
	require.NoError(t, err)

	// Назначаем курьера на заказ
	orderID := "f819526d-6a7c-48eb-b535-43989469d1ca"
	result, err := service.AssignCourier(ctx, orderID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, orderID, result.OrderID)
	assert.Equal(t, courierID, result.CourierID)
	assert.EqualValues(t, modelCourier.TransportCar, result.TransportType)
	assert.False(t, result.Deadline.IsZero())

	// Проверяем, что доставка создана
	delivery, err := deliveryRepository.GetByOrderID(ctx, orderID)
	require.NoError(t, err)
	assert.Equal(t, orderID, delivery.OrderID)
	assert.Equal(t, courierID, delivery.CourierID)
	assert.EqualValues(t, modelDelivery.StatusActive, delivery.Status)

	// Проверяем, что курьер стал занятым
	courier, err := courierRepository.GetByID(ctx, courierID)
	require.NoError(t, err)
	assert.EqualValues(t, modelCourier.StatusBusy, courier.Status)
}

func TestDeliveryService_AssignCourier_OrderAlreadyAssigned(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	deliveryRepository := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	courierRepository := courierRepo.NewCourierRepository(pool)
	transportFactory := deliveryService.NewTransportFactory()
	txManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	fixed := time.Date(2024, 1, 1, 12, 00, 00, 0, time.UTC)
	clock := deliveryService.NewFixedClock(fixed)

	service := deliveryService.NewDeliveryService(
		deliveryRepository,
		courierRepository,
		transportFactory,
		txManager,
		clock,
	)
	ctx := context.Background()

	// Создаем курьера и доставку
	courierData := modelCourier.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        modelCourier.StatusAvailable,
		TransportType: modelCourier.TransportCar,
	}
	courierID, err := courierRepository.Create(ctx, courierData)
	require.NoError(t, err)

	orderID := "f819526d-6a7c-48eb-b535-43989469d1ca"
	deliveryData := modelDelivery.Delivery{
		CourierID:  courierID,
		OrderID:    orderID,
		AssignedAt: time.Now().UTC(),
		Deadline:   time.Now().UTC().Add(30 * time.Minute),
	}
	err = deliveryRepository.Create(ctx, deliveryData)
	require.NoError(t, err)

	// Пытаемся назначить курьера на тот же заказ
	result, err := service.AssignCourier(ctx, orderID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, modelDelivery.ErrOrderAlreadyAssigned)
	assert.Nil(t, result)
}

func TestDeliveryService_AssignCourier_NoAvailableCouriers(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	deliveryRepository := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	courierRepository := courierRepo.NewCourierRepository(pool)
	transportFactory := deliveryService.NewTransportFactory()
	txManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	fixed := time.Date(2024, 1, 1, 12, 00, 00, 0, time.UTC)
	clock := deliveryService.NewFixedClock(fixed)

	service := deliveryService.NewDeliveryService(
		deliveryRepository,
		courierRepository,
		transportFactory,
		txManager,
		clock,
	)
	ctx := context.Background()

	// Создаем только занятого курьера
	courierData := modelCourier.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        modelCourier.StatusBusy,
		TransportType: modelCourier.TransportCar,
	}
	_, err := courierRepository.Create(ctx, courierData)
	require.NoError(t, err)

	// Пытаемся назначить курьера на заказ
	orderID := "f819526d-6a7c-48eb-b535-43989469d1ca"
	result, err := service.AssignCourier(ctx, orderID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, modelCourier.ErrNoAvailableCouriers)
	assert.Nil(t, result)
}

func TestDeliveryService_UnassignCourier_Integration(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	deliveryRepository := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	courierRepository := courierRepo.NewCourierRepository(pool)
	transportFactory := deliveryService.NewTransportFactory()
	txManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	fixed := time.Date(2024, 1, 1, 12, 00, 00, 0, time.UTC)
	clock := deliveryService.NewFixedClock(fixed)

	service := deliveryService.NewDeliveryService(
		deliveryRepository,
		courierRepository,
		transportFactory,
		txManager,
		clock,
	)
	ctx := context.Background()

	// Создаем курьера и доставку
	courierData := modelCourier.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        modelCourier.StatusBusy,
		TransportType: modelCourier.TransportCar,
	}
	courierID, err := courierRepository.Create(ctx, courierData)
	require.NoError(t, err)

	orderID := "f819526d-6a7c-48eb-b535-43989469d1ca"
	deliveryData := modelDelivery.Delivery{
		CourierID:  courierID,
		OrderID:    orderID,
		AssignedAt: time.Now().UTC(),
		Deadline:   time.Now().UTC().Add(30 * time.Minute),
	}
	err = deliveryRepository.Create(ctx, deliveryData)
	require.NoError(t, err)

	// Снимаем курьера с заказа
	result, err := service.UnassignCourier(ctx, orderID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, orderID, result.OrderID)
	assert.Equal(t, courierID, result.CourierID)
	assert.Equal(t, modelDelivery.StatusUnassigned, result.Status)

	// Проверяем, что доставка удалена
	_, err = deliveryRepository.GetByOrderID(ctx, orderID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, modelDelivery.ErrDeliveryNotFound)

	// Проверяем, что курьер стал доступным
	courier, err := courierRepository.GetByID(ctx, courierID)
	require.NoError(t, err)
	assert.EqualValues(t, modelCourier.StatusAvailable, courier.Status)
}

func TestDeliveryService_UnassignCourier_DeliveryNotFound(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	deliveryRepository := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	courierRepository := courierRepo.NewCourierRepository(pool)
	transportFactory := deliveryService.NewTransportFactory()
	txManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	fixed := time.Date(2024, 1, 1, 12, 00, 00, 0, time.UTC)
	clock := deliveryService.NewFixedClock(fixed)

	service := deliveryService.NewDeliveryService(
		deliveryRepository,
		courierRepository,
		transportFactory,
		txManager,
		clock,
	)
	ctx := context.Background()

	// Пытаемся снять курьера с несуществующего заказа
	result, err := service.UnassignCourier(ctx, "c3d7a96d-fbb3-4ea0-a1e6-06023a23b83b")
	assert.Error(t, err)
	assert.ErrorIs(t, err, modelDelivery.ErrDeliveryNotFound)
	assert.Nil(t, result)
}

func TestDeliveryService_Transaction_RollbackOnError(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	deliveryRepository := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	courierRepository := courierRepo.NewCourierRepository(pool)
	transportFactory := deliveryService.NewTransportFactory()
	txManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	fixed := time.Date(2024, 1, 1, 12, 00, 00, 0, time.UTC)
	clock := deliveryService.NewFixedClock(fixed)

	service := deliveryService.NewDeliveryService(
		deliveryRepository,
		courierRepository,
		transportFactory,
		txManager,
		clock,
	)
	ctx := context.Background()

	// Создаем доступного курьера
	courierData := modelCourier.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        modelCourier.StatusAvailable,
		TransportType: modelCourier.TransportCar,
	}
	courierID, err := courierRepository.Create(ctx, courierData)
	require.NoError(t, err)

	// Назначаем курьера на заказ
	orderID := "f819526d-6a7c-48eb-b535-43989469d1ca"
	result1, err := service.AssignCourier(ctx, orderID)
	require.NoError(t, err)
	assert.NotNil(t, result1)

	// Проверяем, что курьер стал занятым
	courier, err := courierRepository.GetByID(ctx, courierID)
	require.NoError(t, err)
	assert.EqualValues(t, modelCourier.StatusBusy, courier.Status)

	// Пытаемся назначить курьера на тот же заказ
	result2, err := service.AssignCourier(ctx, orderID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, modelDelivery.ErrOrderAlreadyAssigned)
	assert.Nil(t, result2)

	// Проверяем, что состояние не изменилось после rollback
	delivery1, err := deliveryRepository.GetByOrderID(ctx, orderID)
	require.NoError(t, err)
	assert.Equal(t, orderID, delivery1.OrderID)
	assert.Equal(t, courierID, delivery1.CourierID)

	// Курьер остался занятым
	courier, err = courierRepository.GetByID(ctx, courierID)
	require.NoError(t, err)
	assert.EqualValues(t, modelCourier.StatusBusy, courier.Status)
}
