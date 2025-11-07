package model

import "time"

type CourierDB struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type CourierUpdateDB struct {
	ID        int64     `json:"id"`
	Name      *string   `json:"name"`
	Phone     *string   `json:"phone"`
	Status    *string   `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}
