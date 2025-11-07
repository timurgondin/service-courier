package model

const (
	CourierStatusPaused    = "paused"
	CourierStatusAvailable = "available"
	CourierStatusBusy      = "busy"
)

var ValidCourierStatuses = []string{
	CourierStatusPaused,
	CourierStatusAvailable,
	CourierStatusBusy,
}
