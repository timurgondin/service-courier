package courier

import "errors"

var (
	ErrCourierNotFound     = errors.New("courier not found")
	ErrPhoneExists         = errors.New("courier with this phone already exists")
	ErrNoAvailableCouriers = errors.New("no available couriers")
)
