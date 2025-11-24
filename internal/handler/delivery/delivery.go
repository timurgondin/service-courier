package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"service-courier/internal/model/delivery"
)

type Handler struct {
	service deliveryService
}

func NewDeliveryHandler(service deliveryService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Assign(w http.ResponseWriter, r *http.Request) {
	var req AssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON",
		})
		return
	}

	if req.OrderID == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON",
		})
		return
	}

	result, err := h.service.AssignCourier(r.Context(), req.OrderID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, ResultToAssignResponse(*result))
}

func (h *Handler) Unassign(w http.ResponseWriter, r *http.Request) {
	var req UnassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON",
		})
		return
	}
	if req.OrderID == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON",
		})
		return
	}
	result, err := h.service.UnassignCourier(r.Context(), req.OrderID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, ResultToUnassignResponse(*result))

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
	case errors.Is(err, delivery.ErrDeliveryNotFound):
		return "Delivery not found", http.StatusNotFound
	case errors.Is(err, delivery.ErrOrderAlreadyAssigned):
		return "Order already assigned", http.StatusConflict
	case errors.Is(err, delivery.ErrNoAvailableCouriers):
		return "No available couriers", http.StatusConflict
	default:
		return "Internal server error", http.StatusInternalServerError
	}
}
