package delivery

import "errors"

var (
	ErrDeliveryNotFound     = errors.New("delivery not found")
	ErrOrderAlreadyAssigned = errors.New("order already assigned")
)
