package delivery_test

//go:generate mockgen -destination=internal/mocks/delivery_repository_mock.go -package=mocks service-courier/internal/service/delivery deliveryRepository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"service-courier/internal/integration"
	modelCourier "service-courier/internal/model/courier"
	modelDelivery "service-courier/internal/model/delivery"
	"service-courier/internal/repository/courier"
	deliveryRepo "service-courier/internal/repository/delivery"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
)

func TestDeliveryRepository_Create(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	repo := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	courierRepo := courier.NewCourierRepository(pool)
	ctx := context.Background()

	// Создаем курьера
	courierData := modelCourier.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        modelCourier.StatusAvailable,
		TransportType: modelCourier.TransportCar,
	}
	courierID, err := courierRepo.Create(ctx, courierData)
	require.NoError(t, err)

	// Создаем доставку
	deliveryData := modelDelivery.Delivery{
		CourierID:  courierID,
		OrderID:    "f819526d-6a7c-48eb-b535-43989469d1ca",
		AssignedAt: time.Now(),
		Deadline:   time.Now().Add(30 * time.Minute),
	}

	err = repo.Create(ctx, deliveryData)
	require.NoError(t, err)

	// Проверяем, что доставка создана
	result, err := repo.GetByOrderID(ctx, "f819526d-6a7c-48eb-b535-43989469d1ca")
	require.NoError(t, err)
	assert.Greater(t, result.ID, int64(0))
	assert.Equal(t, deliveryData.OrderID, result.OrderID)
	assert.Equal(t, deliveryData.CourierID, result.CourierID)
	assert.EqualValues(t, modelDelivery.StatusActive, result.Status)
}

func TestDeliveryRepository_GetByOrderID(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	repo := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	courierRepo := courier.NewCourierRepository(pool)
	ctx := context.Background()

	// Создаем курьера и доставку
	courierData := modelCourier.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        modelCourier.StatusAvailable,
		TransportType: modelCourier.TransportCar,
	}
	courierID, err := courierRepo.Create(ctx, courierData)
	require.NoError(t, err)

	deliveryData := modelDelivery.Delivery{
		CourierID:  courierID,
		OrderID:    "f819526d-6a7c-48eb-b535-43989469d1ca",
		AssignedAt: time.Now(),
		Deadline:   time.Now().Add(30 * time.Minute),
	}
	err = repo.Create(ctx, deliveryData)
	require.NoError(t, err)

	// Получаем доставку
	result, err := repo.GetByOrderID(ctx, "f819526d-6a7c-48eb-b535-43989469d1ca")
	require.NoError(t, err)
	assert.Greater(t, result.ID, int64(0))
	assert.Equal(t, deliveryData.OrderID, result.OrderID)
	assert.Equal(t, deliveryData.CourierID, result.CourierID)
	assert.False(t, result.AssignedAt.IsZero())
	assert.False(t, result.Deadline.IsZero())
}

func TestDeliveryRepository_GetByOrderID_NotFound(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	repo := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	ctx := context.Background()

	result, err := repo.GetByOrderID(ctx, "c3d7a96d-fbb3-4ea0-a1e6-06023a23b83b")
	assert.Error(t, err)
	assert.ErrorIs(t, err, modelDelivery.ErrDeliveryNotFound)
	assert.Nil(t, result)
}

func TestDeliveryRepository_DeleteByOrderID(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	repo := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	courierRepo := courier.NewCourierRepository(pool)
	ctx := context.Background()

	// Создаем курьера и доставку
	courierData := modelCourier.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        modelCourier.StatusAvailable,
		TransportType: modelCourier.TransportCar,
	}
	courierID, err := courierRepo.Create(ctx, courierData)
	require.NoError(t, err)

	deliveryData := modelDelivery.Delivery{
		CourierID:  courierID,
		OrderID:    "f819526d-6a7c-48eb-b535-43989469d1ca",
		AssignedAt: time.Now(),
		Deadline:   time.Now().Add(30 * time.Minute),
	}
	err = repo.Create(ctx, deliveryData)
	require.NoError(t, err)

	// Удаляем доставку
	err = repo.DeleteByOrderID(ctx, "f819526d-6a7c-48eb-b535-43989469d1ca")
	require.NoError(t, err)

	// Проверяем, что доставка удалена
	result, err := repo.GetByOrderID(ctx, "f819526d-6a7c-48eb-b535-43989469d1ca")
	assert.Error(t, err)
	assert.ErrorIs(t, err, modelDelivery.ErrDeliveryNotFound)
	assert.Nil(t, result)
}

func TestDeliveryRepository_DeleteByOrderID_NotFound(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	repo := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	ctx := context.Background()

	err := repo.DeleteByOrderID(ctx, "c3d7a96d-fbb3-4ea0-a1e6-06023a23b83b")
	assert.Error(t, err)
	assert.ErrorIs(t, err, modelDelivery.ErrDeliveryNotFound)
}

func TestDeliveryRepository_ListActiveExpired(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	repo := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	courierRepo := courier.NewCourierRepository(pool)
	ctx := context.Background()

	// Создаем курьера
	courierData := modelCourier.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        modelCourier.StatusAvailable,
		TransportType: modelCourier.TransportCar,
	}
	courierID, err := courierRepo.Create(ctx, courierData)
	require.NoError(t, err)

	// Создаем просроченную доставку (deadline в прошлом)
	baseTime := time.Now().UTC()
	expiredDeadline := baseTime.Add(-1 * time.Hour)
	expiredDelivery := modelDelivery.Delivery{
		CourierID:  courierID,
		OrderID:    "65ae96c6-abff-424b-83fe-92403a4678e1",
		AssignedAt: baseTime.Add(-2 * time.Hour),
		Deadline:   expiredDeadline,
	}
	err = repo.Create(ctx, expiredDelivery)
	require.NoError(t, err)

	// Создаем активную доставку (не просроченную)
	activeDelivery := modelDelivery.Delivery{
		CourierID:  courierID,
		OrderID:    "72e9c6de-f88b-47b3-968d-0d9bff7af1e1",
		AssignedAt: baseTime,
		Deadline:   baseTime.Add(30 * time.Minute),
	}
	err = repo.Create(ctx, activeDelivery)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	checkTime := time.Now().UTC()

	// Получаем просроченные доставки
	expired, err := repo.ListActiveExpired(ctx, checkTime)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(expired), 1)

	// Проверяем, что найдена просроченная доставка
	found := false
	for _, d := range expired {
		if d.OrderID == "65ae96c6-abff-424b-83fe-92403a4678e1" {
			found = true
			assert.EqualValues(t, modelDelivery.StatusActive, d.Status)
			assert.True(t, d.Deadline.Before(checkTime.Add(1*time.Second)),
				"deadline should be before checkTime, got deadline=%v, checkTime=%v", d.Deadline, checkTime)
		}
	}
	assert.True(t, found, "expired delivery should be found")

	// Проверяем, что не просроченная доставка не попала в список
	for _, d := range expired {
		assert.NotEqual(t, "72e9c6de-f88b-47b3-968d-0d9bff7af1e1", d.OrderID, "active delivery should not be in expired list")
	}
}

func TestDeliveryRepository_ListActiveExpired_NoExpired(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	repo := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	courierRepo := courier.NewCourierRepository(pool)
	ctx := context.Background()

	// Создаем курьера
	courierData := modelCourier.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        modelCourier.StatusAvailable,
		TransportType: modelCourier.TransportCar,
	}
	courierID, err := courierRepo.Create(ctx, courierData)
	require.NoError(t, err)

	now := time.Now()

	// Создаем только активную доставку (не просроченную)
	activeDelivery := modelDelivery.Delivery{
		CourierID:  courierID,
		OrderID:    "72e9c6de-f88b-47b3-968d-0d9bff7af1e1",
		AssignedAt: now,
		Deadline:   now.Add(30 * time.Minute),
	}
	err = repo.Create(ctx, activeDelivery)
	require.NoError(t, err)

	// Получаем просроченные доставки
	expired, err := repo.ListActiveExpired(ctx, now)
	require.NoError(t, err)
	assert.Equal(t, 0, len(expired), "should be no expired deliveries")
}

func TestDeliveryRepository_UpdateStatusByIDs(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	repo := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	courierRepo := courier.NewCourierRepository(pool)
	ctx := context.Background()

	// Создаем курьера
	courierData := modelCourier.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        modelCourier.StatusAvailable,
		TransportType: modelCourier.TransportCar,
	}
	courierID, err := courierRepo.Create(ctx, courierData)
	require.NoError(t, err)

	now := time.Now()

	// Создаем несколько доставок
	delivery1 := modelDelivery.Delivery{
		CourierID:  courierID,
		OrderID:    "f819526d-6a7c-48eb-b535-43989469d1ca",
		AssignedAt: now,
		Deadline:   now.Add(30 * time.Minute),
	}
	err = repo.Create(ctx, delivery1)
	require.NoError(t, err)

	// Получаем ID из созданной доставки
	created1, err := repo.GetByOrderID(ctx, "f819526d-6a7c-48eb-b535-43989469d1ca")
	require.NoError(t, err)
	delivery1ID := created1.ID
	require.Greater(t, delivery1ID, int64(0), "delivery1.ID should be set")

	delivery2 := modelDelivery.Delivery{
		CourierID:  courierID,
		OrderID:    "e00b99da-4812-4401-8f54-af2cba66b819",
		AssignedAt: now,
		Deadline:   now.Add(30 * time.Minute),
	}
	err = repo.Create(ctx, delivery2)
	require.NoError(t, err)

	// Получаем ID из созданной доставки
	created2, err := repo.GetByOrderID(ctx, "e00b99da-4812-4401-8f54-af2cba66b819")
	require.NoError(t, err)
	delivery2ID := created2.ID
	require.Greater(t, delivery2ID, int64(0), "delivery2.ID should be set")

	// Обновляем статусы
	err = repo.UpdateStatusByIDs(ctx, []int64{delivery1ID, delivery2ID}, modelDelivery.StatusCompleted)
	require.NoError(t, err)

	// Проверяем обновление
	result1, err := repo.GetByOrderID(ctx, "f819526d-6a7c-48eb-b535-43989469d1ca")
	require.NoError(t, err)
	assert.Equal(t, string(modelDelivery.StatusCompleted), string(result1.Status))

	result2, err := repo.GetByOrderID(ctx, "e00b99da-4812-4401-8f54-af2cba66b819")
	require.NoError(t, err)
	assert.Equal(t, string(modelDelivery.StatusCompleted), string(result2.Status))
}

func TestDeliveryRepository_UpdateStatusByIDs_EmptySlice(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	ctxGetter := trmpgx.DefaultCtxGetter
	repo := deliveryRepo.NewDeliveryRepository(pool, ctxGetter)
	ctx := context.Background()

	// Обновляем статусы пустого списка (должно вернуть nil)
	err := repo.UpdateStatusByIDs(ctx, []int64{}, modelDelivery.StatusCompleted)
	require.NoError(t, err)
}
