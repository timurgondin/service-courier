package courier

import "time"

type Courier struct {
	ID        int64
	Name      string
	Phone     string
	Status    CourierStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CourierStatus string

const (
	StatusAvailable = "available"
	StatusBusy      = "busy"
	StatusPaused    = "paused"
)
