package order

import "time"

type Order struct {
	ID        string
	Status    string
	CreatedAt time.Time
}

const (
	StatusCreated   = "created"
	StatusCancelled = "cancelled"
	StatusCompleted = "completed"
)
