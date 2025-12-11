package courier_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	courierService "service-courier/internal/service/courier"
	courierRepo "service-courier/internal/repository/courier"
	"service-courier/internal/integration"
	model "service-courier/internal/model/courier"
)

func TestCourierService_Integration(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	service := courierService.NewCourierService(repo)
	ctx := context.Background()

	// Тест создания курьера
	courierData := model.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusAvailable,
		TransportType: model.TransportCar,
	}

	id, err := service.CreateCourier(ctx, courierData)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Тест получения курьера
	result, err := service.GetCourier(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, courierData.Name, result.Name)
	assert.Equal(t, courierData.Phone, result.Phone)
	assert.EqualValues(t, courierData.Status, result.Status)
	assert.EqualValues(t, courierData.TransportType, result.TransportType)

	// Тест обновления курьера
	updatedData := model.Courier{
		ID:     id,
		Name:   "Ivan Updated",
		Status: model.StatusBusy,
	}
	err = service.UpdateCourier(ctx, updatedData)
	require.NoError(t, err)

	// Проверяем обновление
	updated, err := service.GetCourier(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "Ivan Updated", updated.Name)
	assert.EqualValues(t, model.StatusBusy, updated.Status)
	// Phone и TransportType не должны измениться
	assert.Equal(t, courierData.Phone, updated.Phone)
	assert.EqualValues(t, courierData.TransportType, updated.TransportType)

	// Тест получения всех курьеров
	allCouriers, err := service.GetAllCouriers(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(allCouriers), 1)

	// Проверяем, что созданный курьер есть в списке
	found := false
	for _, c := range allCouriers {
		if c.ID == id {
			found = true
			assert.Equal(t, "Ivan Updated", c.Name)
			assert.EqualValues(t, model.StatusBusy, c.Status)
		}
	}
	assert.True(t, found, "created courier should be in the list")
}

func TestCourierService_Integration_PhoneExists(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	service := courierService.NewCourierService(repo)
	ctx := context.Background()

	// Создаем первого курьера
	courierData := model.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusAvailable,
		TransportType: model.TransportCar,
	}

	id, err := service.CreateCourier(ctx, courierData)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Пытаемся создать второго курьера с тем же телефоном
	_, err = service.CreateCourier(ctx, courierData)
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrPhoneExists)
}

func TestCourierService_Integration_NotFound(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	service := courierService.NewCourierService(repo)
	ctx := context.Background()

	// Пытаемся получить несуществующего курьера
	result, err := service.GetCourier(ctx, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrCourierNotFound)
	assert.Nil(t, result)

	// Пытаемся обновить несуществующего курьера
	updatedData := model.Courier{
		ID:     99999,
		Name:   "Non-existent",
		Status: model.StatusAvailable,
	}
	err = service.UpdateCourier(ctx, updatedData)
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrCourierNotFound)
}
