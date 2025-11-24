package delivery

import (
	"service-courier/internal/service/delivery"
	"time"
)

func ResultToAssignResponse(res delivery.AssignResult) AssignResponse {
	return AssignResponse{
		CourierID:        res.CourierID,
		OrderID:          res.OrderID,
		TransportType:    string(res.TransportType),
		DeliveryDeadline: res.DeliveryDeadline.Format(time.RFC3339),
	}
}

func ResultToUnassignResponse(res delivery.UnassignResult) UnassignResponse {
	return UnassignResponse{
		OrderID:   res.OrderID,
		Status:    res.Status,
		CourierID: res.CourierID,
	}
}
