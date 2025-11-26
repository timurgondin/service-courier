package delivery

// AssignRequest запрос на назначение курьера на заказ
type AssignRequest struct {
	OrderID string `json:"order_id"`
}

// AssignResponse ответ на назначение курьера
type AssignResponse struct {
	CourierID     int64  `json:"courier_id"`
	OrderID       string `json:"order_id"`
	TransportType string `json:"transport_type"`
	Deadline      string `json:"delivery_deadline"`
}

// UnassignRequest запрос на снятие курьера с заказа
type UnassignRequest struct {
	OrderID string `json:"order_id"`
}

// UnassignResponse ответ на снятие курьера
type UnassignResponse struct {
	OrderID   string `json:"order_id"`
	Status    string `json:"status"`
	CourierID int64  `json:"courier_id"`
}
