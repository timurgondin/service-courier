package delivery

import (
	"service-courier/internal/model/courier"
	"time"
)

type DeliveryTimeFactory struct{}

func NewDeliveryTimeFactory() *DeliveryTimeFactory {
	return &DeliveryTimeFactory{}
}

func (f *DeliveryTimeFactory) CalculateDeadline(transportType courier.TransportType, baseTime time.Time) time.Time {
	var duration time.Duration

	switch transportType {
	case courier.TransportOnFoot:
		duration = 30 * time.Minute
	case courier.TransportScooter:
		duration = 15 * time.Minute
	case courier.TransportCar:
		duration = 5 * time.Minute
	}

	return baseTime.Add(duration)
}
