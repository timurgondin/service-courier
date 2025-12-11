package delivery

import "time"

type Delivery struct {
	ID         int64
	CourierID  int64
	OrderID    string
	Status     DeliveryStatus
	AssignedAt time.Time
	Deadline   time.Time
}

type DeliveryStatus string

const (
	StatusActive    = "active"
	StatusCompleted = "completed"
	StatusDeleted   = "deleted"
)

const StatusUnassigned = "unassigned"
