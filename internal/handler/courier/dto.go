package courier

// Courier - модель курьера для ответа
type Courier struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

// CreateRequest запрос на создание курьера
type CreateRequest struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

// UpdateRequest запрос на обновление курьера
type UpdateRequest Courier
