package courier_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"service-courier/internal/integration"
	model "service-courier/internal/model/courier"
	courierRepo "service-courier/internal/repository/courier"
)

func TestCourierRepository_Create(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	ctx := context.Background()

	courierData := model.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusAvailable,
		TransportType: model.TransportCar,
	}

	id, err := repo.Create(ctx, courierData)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Проверяем, что курьер создан
	created, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, courierData.Name, created.Name)
	assert.Equal(t, courierData.Phone, created.Phone)
	assert.EqualValues(t, courierData.Status, created.Status)
	assert.EqualValues(t, courierData.TransportType, created.TransportType)
}

func TestCourierRepository_Create_DuplicatePhone(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	ctx := context.Background()

	courierData := model.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusAvailable,
		TransportType: model.TransportCar,
	}

	id, err := repo.Create(ctx, courierData)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Пытаемся создать курьера с тем же телефоном
	_, err = repo.Create(ctx, courierData)
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrPhoneExists)
}

func TestCourierRepository_GetByID(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	ctx := context.Background()

	// Создаем курьера
	courierData := model.Courier{
		Name:          "Petr",
		Phone:         "+78005553536",
		Status:        model.StatusBusy,
		TransportType: model.TransportScooter,
	}

	id, err := repo.Create(ctx, courierData)
	require.NoError(t, err)

	// Получаем курьера
	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, courierData.Name, result.Name)
	assert.Equal(t, courierData.Phone, result.Phone)
	assert.EqualValues(t, courierData.Status, result.Status)
	assert.EqualValues(t, courierData.TransportType, result.TransportType)
	assert.False(t, result.CreatedAt.IsZero())
	assert.False(t, result.UpdatedAt.IsZero())
}

func TestCourierRepository_GetByID_NotFound(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	ctx := context.Background()

	result, err := repo.GetByID(ctx, 99999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrCourierNotFound)
	assert.Nil(t, result)
}

func TestCourierRepository_GetAll(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	ctx := context.Background()

	// Создаем несколько курьеров
	courier1 := model.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusAvailable,
		TransportType: model.TransportCar,
	}
	courier2 := model.Courier{
		Name:          "Petr",
		Phone:         "+78005553536",
		Status:        model.StatusBusy,
		TransportType: model.TransportScooter,
	}

	id1, err := repo.Create(ctx, courier1)
	require.NoError(t, err)

	id2, err := repo.Create(ctx, courier2)
	require.NoError(t, err)

	// Получаем всех курьеров
	results, err := repo.GetAll(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)

	// Проверяем, что созданные курьеры есть в списке
	found1 := false
	found2 := false
	for _, c := range results {
		if c.ID == id1 {
			found1 = true
			assert.Equal(t, courier1.Name, c.Name)
		}
		if c.ID == id2 {
			found2 = true
			assert.Equal(t, courier2.Name, c.Name)
		}
	}
	assert.True(t, found1, "courier 1 should be found")
	assert.True(t, found2, "courier 2 should be found")
}

func TestCourierRepository_Update(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	ctx := context.Background()

	// Создаем курьера
	courierData := model.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusAvailable,
		TransportType: model.TransportCar,
	}

	id, err := repo.Create(ctx, courierData)
	require.NoError(t, err)

	// Обновляем курьера
	updatedData := model.Courier{
		ID:     id,
		Name:   "Ivan Updated",
		Status: model.StatusBusy,
	}

	err = repo.Update(ctx, updatedData)
	require.NoError(t, err)

	// Проверяем обновление
	result, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "Ivan Updated", result.Name)
	assert.EqualValues(t, model.StatusBusy, result.Status)
	// Phone и TransportType не должны измениться
	assert.Equal(t, courierData.Phone, result.Phone)
	assert.EqualValues(t, courierData.TransportType, result.TransportType)
}

func TestCourierRepository_Update_NotFound(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	ctx := context.Background()

	updatedData := model.Courier{
		ID:     99999,
		Name:   "Non-existent",
		Status: model.StatusAvailable,
	}

	err := repo.Update(ctx, updatedData)
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrCourierNotFound)
}

func TestCourierRepository_GetAvailableWithMinDeliveries(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	ctx := context.Background()

	// Создаем доступного курьера
	courier1 := model.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusAvailable,
		TransportType: model.TransportCar,
	}
	id1, err := repo.Create(ctx, courier1)
	require.NoError(t, err)

	// Создаем еще одного доступного курьера
	courier2 := model.Courier{
		Name:          "Petr",
		Phone:         "+78005553536",
		Status:        model.StatusAvailable,
		TransportType: model.TransportScooter,
	}
	id2, err := repo.Create(ctx, courier2)
	require.NoError(t, err)

	// Получаем доступного курьера с минимальным количеством доставок
	result, err := repo.GetAvailableWithMinDeliveries(ctx)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.ID == id1 || result.ID == id2)
	assert.EqualValues(t, model.StatusAvailable, result.Status)
}

func TestCourierRepository_GetAvailableWithMinDeliveries_NoAvailable(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	ctx := context.Background()

	// Создаем только занятого курьера
	courier := model.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusBusy,
		TransportType: model.TransportCar,
	}
	_, err := repo.Create(ctx, courier)
	require.NoError(t, err)

	// Пытаемся получить доступного курьера
	result, err := repo.GetAvailableWithMinDeliveries(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrNoAvailableCouriers)
	assert.Nil(t, result)
}

func TestCourierRepository_UpdateStatusBatch(t *testing.T) {
	pool, cleanup := integration.SetupTestDB(t)
	defer cleanup()

	repo := courierRepo.NewCourierRepository(pool)
	ctx := context.Background()

	// Создаем несколько курьеров
	courier1 := model.Courier{
		Name:          "Ivan",
		Phone:         "+78005553535",
		Status:        model.StatusBusy,
		TransportType: model.TransportCar,
	}
	courier2 := model.Courier{
		Name:          "Petr",
		Phone:         "+78005553536",
		Status:        model.StatusBusy,
		TransportType: model.TransportScooter,
	}

	id1, err := repo.Create(ctx, courier1)
	require.NoError(t, err)

	id2, err := repo.Create(ctx, courier2)
	require.NoError(t, err)

	// Обновляем статусы массово
	err = repo.UpdateStatusBatch(ctx, []int64{id1, id2}, model.StatusAvailable)
	require.NoError(t, err)

	// Проверяем обновление
	result1, err := repo.GetByID(ctx, id1)
	require.NoError(t, err)
	assert.EqualValues(t, model.StatusAvailable, result1.Status)

	result2, err := repo.GetByID(ctx, id2)
	require.NoError(t, err)
	assert.EqualValues(t, model.StatusAvailable, result2.Status)
}
