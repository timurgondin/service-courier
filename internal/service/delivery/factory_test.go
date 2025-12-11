package delivery_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	modelCourier "service-courier/internal/model/courier"
	deliveryService "service-courier/internal/service/delivery"
)

func TestDeliveryTimeFactory_CalculateDeadline_OnFoot(t *testing.T) {
	t.Parallel()

	factory := deliveryService.NewTransportFactory()
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	transport := factory.Create(modelCourier.TransportOnFoot)
	deadline := baseTime.Add(transport.DeliveryDuration())

	expected := baseTime.Add(30 * time.Minute)
	assert.Equal(t, expected, deadline)
}

func TestDeliveryTimeFactory_CalculateDeadline_Scooter(t *testing.T) {
	t.Parallel()

	factory := deliveryService.NewTransportFactory()
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	transport := factory.Create(modelCourier.TransportScooter)
	deadline := baseTime.Add(transport.DeliveryDuration())

	expected := baseTime.Add(15 * time.Minute)
	assert.Equal(t, expected, deadline)
}

func TestDeliveryTimeFactory_CalculateDeadline_Car(t *testing.T) {
	t.Parallel()

	factory := deliveryService.NewTransportFactory()
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	transport := factory.Create(modelCourier.TransportCar)
	deadline := baseTime.Add(transport.DeliveryDuration())

	expected := baseTime.Add(5 * time.Minute)
	assert.Equal(t, expected, deadline)
}
