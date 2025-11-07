package model

import "time"

type Courier struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type CourierCreateRequest struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

type CourierUpdateRequest struct {
	ID     int64   `json:"id"`
	Name   *string `json:"name"`
	Phone  *string `json:"phone"`
	Status *string `json:"status"`
}
