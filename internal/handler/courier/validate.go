package courier

import (
	"fmt"
	"service-courier/internal/model/courier"
	"unicode"
)

func (r CreateRequest) Validate() error {
	if err := validateName(r.Name); err != nil {
		return err
	}
	if err := validatePhone(r.Phone); err != nil {
		return err
	}
	if err := validateStatus(r.Status); err != nil {
		return err
	}
	if err := validateTransportType(r.TransportType); err != nil {
		return err
	}
	return nil
}

func (r UpdateRequest) Validate() error {
	if r.ID <= 0 {
		return fmt.Errorf("Invalid id")
	}

	if r.Name == "" && r.Phone == "" && r.Status == "" && r.TransportType == "" {
		return fmt.Errorf("All fields are empty")
	}

	if r.Name != "" {
		if err := validateName(r.Name); err != nil {
			return err
		}
	}
	if r.Phone != "" {
		if err := validatePhone(r.Phone); err != nil {
			return err
		}
	}
	if r.Status != "" {
		if err := validateStatus(r.Status); err != nil {
			return err
		}
	}
	if r.TransportType != "" {
		if err := validateTransportType(r.TransportType); err != nil {
			return err
		}
	}

	return nil
}

func validateName(name string) error {
	if name == "" || len(name) > 100 {
		return fmt.Errorf("Name is too long or empty")
	}
	return nil
}

func validatePhone(phone string) error {
	if phone == "" {
		return fmt.Errorf("Phone is empty")
	}

	if len(phone) != 12 || phone[0] != '+' {
		return fmt.Errorf("Invalid phone")
	}

	digits := phone[1:]
	for _, digit := range digits {
		if !unicode.IsDigit(digit) {
			return fmt.Errorf("Invalid phone")
		}
	}

	return nil
}

func validateStatus(status string) error {
	switch status {
	case courier.StatusAvailable, courier.StatusBusy, courier.StatusPaused:
		return nil
	default:
		return fmt.Errorf("Invalid status")
	}
}

func validateTransportType(transportType string) error {
	switch transportType {
	case courier.TransportOnFoot, courier.TransportScooter, courier.TransportCar:
		return nil
	default:
		return fmt.Errorf("Invalid transport type")
	}
}
