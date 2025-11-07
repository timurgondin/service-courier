package handler

import (
	"encoding/json"
	"net/http"
	"service-courier/internal/model"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CourierHandler struct {
	service CourierService
}

func NewCourierHandler(service CourierService) *CourierHandler {
	return &CourierHandler{service: service}
}

func (h *CourierHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error": "Invalid courier ID"}`, http.StatusBadRequest)
		return
	}

	var courier *model.Courier
	courier, err = h.service.GetCourier(r.Context(), id)

	if err != nil {
		switch err {
		case model.ErrCourierNotFound:
			http.Error(w, `{"error": "Courier not found"}`, http.StatusNotFound)
		default:
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
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
		http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(couriers)
}

func (h *CourierHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CourierCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	id, err := h.service.CreateCourier(r.Context(), &req)

	if err != nil {
		switch err {
		case model.ErrCourierNameEmpty:
			http.Error(w, `{"error": "Courier name is required"}`, http.StatusBadRequest)
		case model.ErrCourierPhoneEmpty:
			http.Error(w, `{"error": "Courier phone is required"}`, http.StatusBadRequest)
		case model.ErrCourierPhoneInvalid:
			http.Error(w, `{"error": "Courier phone must start with '+', have 12 characters, and contain only digits after '+'"}`, http.StatusBadRequest)
		case model.ErrPhoneExists:
			http.Error(w, `{"error": "Courier with this phone already exists"}`, http.StatusConflict)
		case model.ErrCourierStatusEmpty:
			http.Error(w, `{"error": "Courier status is required"}`, http.StatusBadRequest)
		case model.ErrCourierStatusInvalid:
			http.Error(w, `{"error": "Courier status must be one of: paused, available, busy"}`, http.StatusBadRequest)
		default:
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
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
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	err := h.service.UpdateCourier(r.Context(), &req)

	if err != nil {
		switch err {
		case model.ErrCourierNotFound:
			http.Error(w, `{"error": "Courier not found"}`, http.StatusNotFound)
		case model.ErrCourierNameEmpty:
			http.Error(w, `{"error": "Courier name is required"}`, http.StatusBadRequest)
		case model.ErrCourierPhoneEmpty:
			http.Error(w, `{"error": "Courier phone is required"}`, http.StatusBadRequest)
		case model.ErrCourierPhoneInvalid:
			http.Error(w, `{"error": "Courier phone must start with '+', have 12 characters, and contain only digits after '+'"}`, http.StatusBadRequest)
		case model.ErrCourierStatusEmpty:
			http.Error(w, `{"error": "Courier status is required"}`, http.StatusBadRequest)
		case model.ErrCourierStatusInvalid:
			http.Error(w, `{"error": "Courier status must be one of: paused, available, busy"}`, http.StatusBadRequest)
		default:
			http.Error(w, `{"error": "Database error"}`, http.StatusInternalServerError)
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
