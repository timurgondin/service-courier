package courier

import (
	"encoding/json"
	"errors"
	"net/http"
	"service-courier/internal/model/courier"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service courierService
}

func NewCourierHandler(service courierService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid courier ID",
		})
		return
	}

	courierData, err := h.service.GetCourier(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, ModelToResponse(*courierData))
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	couriers, err := h.service.GetAllCouriers(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}

	responseCouriers := make([]Courier, len(couriers))
	for i, c := range couriers {
		responseCouriers[i] = ModelToResponse(c)
	}

	h.writeJSON(w, http.StatusOK, responseCouriers)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON",
		})
		return
	}

	if err := req.Validate(); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	id, err := h.service.CreateCourier(r.Context(), req.ToModel())
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      id,
		"message": "Courier created successfully",
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON",
		})
		return
	}

	if err := req.Validate(); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	err := h.service.UpdateCourier(r.Context(), req.ToModel())
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"message": "Courier updated successfully",
	})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	message, status := h.mapError(err)
	h.writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) mapError(err error) (string, int) {
	switch {
	case errors.Is(err, courier.ErrCourierNotFound):
		return "Courier not found", http.StatusNotFound
	case errors.Is(err, courier.ErrPhoneExists):
		return "Courier with this phone already exists", http.StatusConflict
	default:
		return "Internal server error", http.StatusInternalServerError
	}
}
