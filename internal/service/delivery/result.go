package delivery

import (
	"service-courier/internal/model/courier"
	"time"
)

type AssignResult struct {
	CourierID        int64
	OrderID          string
	TransportType    courier.TransportType
	DeliveryDeadline time.Time
}

type UnassignResult struct {
	OrderID   string
	Status    string
	CourierID int64
}
