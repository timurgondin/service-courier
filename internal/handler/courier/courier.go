package courier

import (
	"encoding/json"
	"errors"
	"net/http"
	"service-courier/internal/model"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CourierHandler struct {
	service courierService
}

func NewCourierHandler(service courierService) *CourierHandler {
	return &CourierHandler{service: service}
}

func (h *CourierHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid courier ID")
		return
	}

	var courier *model.Courier
	courier, err = h.service.GetCourier(r.Context(), id)

	if err != nil {
		switch err {
		case model.ErrCourierNotFound:
			writeJSONError(w, http.StatusNotFound, "Courier not found")
		default:
			writeJSONError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(courier)
}

func (h *CourierHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	couriers, err := h.service.GetAllCouriers(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(couriers)
}

func (h *CourierHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CourierCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	id, err := h.service.CreateCourier(r.Context(), &req)

	if err != nil {
		switch {
		case errors.Is(err, model.ErrCourierNameEmpty):
			writeJSONError(w, http.StatusBadRequest, "Courier name is required")
		case errors.Is(err, model.ErrCourierPhoneEmpty):
			writeJSONError(w, http.StatusBadRequest, "Courier phone is required")
		case errors.Is(err, model.ErrCourierPhoneInvalid):
			writeJSONError(w, http.StatusBadRequest, "Courier phone must start with '+', have 12 characters, and contain only digits after '+'")
		case errors.Is(err, model.ErrPhoneExists):
			writeJSONError(w, http.StatusConflict, "Courier with this phone already exists")
		case errors.Is(err, model.ErrCourierStatusEmpty):
			writeJSONError(w, http.StatusBadRequest, "Courier status is required")
		case errors.Is(err, model.ErrCourierStatusInvalid):
			writeJSONError(w, http.StatusBadRequest, "Courier status must be one of: paused, available, busy")
		default:
			writeJSONError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	response := map[string]interface{}{
		"id":      id,
		"message": "Profile created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *CourierHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req model.CourierUpdateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	err := h.service.UpdateCourier(r.Context(), &req)

	if err != nil {
		switch {
		case errors.Is(err, model.ErrCourierNotFound):
			writeJSONError(w, http.StatusNotFound, "Courier not found")
		case errors.Is(err, model.ErrCourierNameEmpty):
			writeJSONError(w, http.StatusBadRequest, "Courier name is required")
		case errors.Is(err, model.ErrCourierPhoneEmpty):
			writeJSONError(w, http.StatusBadRequest, "Courier phone is required")
		case errors.Is(err, model.ErrPhoneExists):
			writeJSONError(w, http.StatusConflict, "Courier with this phone already exists")
		case errors.Is(err, model.ErrCourierPhoneInvalid):
			writeJSONError(w, http.StatusBadRequest, "Courier phone must start with '+', have 12 characters, and contain only digits after '+'")
		case errors.Is(err, model.ErrCourierStatusEmpty):
			writeJSONError(w, http.StatusBadRequest, "Courier status is required")
		case errors.Is(err, model.ErrCourierStatusInvalid):
			writeJSONError(w, http.StatusBadRequest, "Courier status must be one of: paused, available, busy")
		default:
			writeJSONError(w, http.StatusInternalServerError, "Database error")
		}
		return
	}

	response := map[string]string{
		"message": "Profile updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
