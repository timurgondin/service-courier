package delivery

import "time"

type Transport interface {
	DeliveryDuration() time.Duration
}

type OnFoot struct{}

func (OnFoot) DeliveryDuration() time.Duration {
	return 30 * time.Minute
}

type Scooter struct{}

func (Scooter) DeliveryDuration() time.Duration {
	return 15 * time.Minute
}

type Car struct{}

func (Car) DeliveryDuration() time.Duration {
	return 5 * time.Minute
}
