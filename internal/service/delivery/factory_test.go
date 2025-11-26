package delivery_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	deliveryService "service-courier/internal/service/delivery"
	modelCourier "service-courier/internal/model/courier"
)

func TestDeliveryTimeFactory_CalculateDeadline_OnFoot(t *testing.T) {
	t.Parallel()

	factory := deliveryService.NewDeliveryTimeFactory()
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	deadline := factory.CalculateDeadline(modelCourier.TransportOnFoot, baseTime)

	expected := baseTime.Add(30 * time.Minute)
	assert.Equal(t, expected, deadline)
}

func TestDeliveryTimeFactory_CalculateDeadline_Scooter(t *testing.T) {
	t.Parallel()

	factory := deliveryService.NewDeliveryTimeFactory()
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	deadline := factory.CalculateDeadline(modelCourier.TransportScooter, baseTime)

	expected := baseTime.Add(15 * time.Minute)
	assert.Equal(t, expected, deadline)
}

func TestDeliveryTimeFactory_CalculateDeadline_Car(t *testing.T) {
	t.Parallel()

	factory := deliveryService.NewDeliveryTimeFactory()
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	deadline := factory.CalculateDeadline(modelCourier.TransportCar, baseTime)

	expected := baseTime.Add(5 * time.Minute)
	assert.Equal(t, expected, deadline)
}
