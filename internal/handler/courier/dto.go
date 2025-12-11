package courier

// Courier - модель курьера для ответа
type Courier struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Status        string `json:"status"`
	TransportType string `json:"transport_type"`
}

// CreateRequest запрос на создание курьера
type CreateRequest struct {
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Status        string `json:"status"`
	TransportType string `json:"transport_type"`
}

// UpdateRequest запрос на обновление курьера
type UpdateRequest Courier
