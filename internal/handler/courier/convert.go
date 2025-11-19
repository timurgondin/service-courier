package courier

import "service-courier/internal/model/courier"

func ModelToResponse(courier courier.Courier) Courier {
	return Courier{
		ID:     courier.ID,
		Name:   courier.Name,
		Phone:  courier.Phone,
		Status: string(courier.Status),
	}
}

func (r CreateRequest) ToModel() courier.Courier {
	return courier.Courier{
		Name:   r.Name,
		Phone:  r.Phone,
		Status: courier.CourierStatus(r.Status),
	}
}

func (r UpdateRequest) ToModel() courier.Courier {
	return courier.Courier{
		ID:     r.ID,
		Name:   r.Name,
		Phone:  r.Phone,
		Status: courier.CourierStatus(r.Status),
	}
}
