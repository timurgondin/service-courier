package model

import "time"

type CourierDB struct {
	ID        int64
	Name      string
	Phone     string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CourierUpdateDB struct {
	ID        int64
	Name      *string
	Phone     *string
	Status    *string
	UpdatedAt time.Time
}
