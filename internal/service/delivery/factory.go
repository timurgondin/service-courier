package delivery

import (
	"service-courier/internal/model/courier"
)

type TransportFactory interface {
	Create(t courier.TransportType) Transport
}

type DefaultTransportFactory struct{}

func NewTransportFactory() TransportFactory {
	return &DefaultTransportFactory{}
}

func (f *DefaultTransportFactory) Create(t courier.TransportType) Transport {
	switch t {
	case courier.TransportOnFoot:
		return OnFoot{}
	case courier.TransportScooter:
		return Scooter{}
	case courier.TransportCar:
		return Car{}
	default:
		return nil
	}
}
