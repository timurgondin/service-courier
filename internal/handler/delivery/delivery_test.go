package delivery_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	deliveryHandler "service-courier/internal/handler/delivery"
	"service-courier/internal/mocks"
	modelCourier "service-courier/internal/model/courier"
	modelDelivery "service-courier/internal/model/delivery"
	dtoDelivery "service-courier/internal/service/delivery"
)

/************ TEST: POST /delivery/assign ************/

func TestAssignCourier_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockdeliveryService(ctrl)

	mockService.EXPECT().
		AssignCourier(gomock.Any(), "order-1").
		Return(&dtoDelivery.AssignResult{
			OrderID:       "order-1",
			CourierID:     10,
			TransportType: "car",
			Deadline:      time.Date(2025, 11, 30, 12, 0, 0, 0, time.UTC),
		}, nil)

	h := deliveryHandler.NewDeliveryHandler(mockService)
	r := chi.NewRouter()
	r.Post("/delivery/assign", h.Assign)

	body := `{"order_id":"order-1"}`
	req := httptest.NewRequest("POST", "/delivery/assign", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	// Проверяем поля DTO
	var resp deliveryHandler.AssignResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.CourierID != 10 {
		t.Fatalf("expected courier_id=10, got %d", resp.CourierID)
	}
	if resp.OrderID != "order-1" {
		t.Fatalf("expected order_id='order-1', got %s", resp.OrderID)
	}
	if resp.TransportType != "car" {
		t.Fatalf("expected transport_type='car', got %s", resp.TransportType)
	}
	if resp.Deadline != "2025-11-30T12:00:00Z" {
		t.Fatalf("expected delivery_deadline='2025-11-30T12:00:00Z', got %s", resp.Deadline)
	}
}

func TestAssignCourier_InvalidJSON(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockdeliveryService(ctrl)
	h := deliveryHandler.NewDeliveryHandler(mockService)
	r := chi.NewRouter()
	r.Post("/delivery/assign", h.Assign)

	req := httptest.NewRequest("POST", "/delivery/assign", bytes.NewBuffer([]byte("{invalid")))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestAssignCourier_EmptyOrderID(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockdeliveryService(ctrl)
	h := deliveryHandler.NewDeliveryHandler(mockService)
	r := chi.NewRouter()
	r.Post("/delivery/assign", h.Assign)

	body := `{"order_id":""}`
	req := httptest.NewRequest("POST", "/delivery/assign", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestAssignCourier_OrderAlreadyAssigned(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockdeliveryService(ctrl)
	mockService.EXPECT().
		AssignCourier(gomock.Any(), "order-1").
		Return(nil, modelDelivery.ErrOrderAlreadyAssigned)

	h := deliveryHandler.NewDeliveryHandler(mockService)
	r := chi.NewRouter()
	r.Post("/delivery/assign", h.Assign)

	body := `{"order_id":"order-1"}`
	req := httptest.NewRequest("POST", "/delivery/assign", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d", rr.Code)
	}
}

func TestAssignCourier_NoAvailableCouriers(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockdeliveryService(ctrl)
	mockService.EXPECT().
		AssignCourier(gomock.Any(), "order-1").
		Return(nil, modelCourier.ErrNoAvailableCouriers)

	h := deliveryHandler.NewDeliveryHandler(mockService)
	r := chi.NewRouter()
	r.Post("/delivery/assign", h.Assign)

	body := `{"order_id":"order-1"}`
	req := httptest.NewRequest("POST", "/delivery/assign", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d", rr.Code)
	}
}

/************ TEST: POST /delivery/unassign ************/

func TestUnassignCourier_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockdeliveryService(ctrl)
	mockService.EXPECT().
		UnassignCourier(gomock.Any(), "order-1").
		Return(&dtoDelivery.UnassignResult{
			OrderID:   "order-1",
			CourierID: 10,
			Status:    "unassigned",
		}, nil)

	h := deliveryHandler.NewDeliveryHandler(mockService)
	r := chi.NewRouter()
	r.Post("/delivery/unassign", h.Unassign)

	body := `{"order_id":"order-1"}`
	req := httptest.NewRequest("POST", "/delivery/unassign", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}

	var resp deliveryHandler.UnassignResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.CourierID != 10 {
		t.Fatalf("expected courier_id=10, got %d", resp.CourierID)
	}
	if resp.OrderID != "order-1" {
		t.Fatalf("expected order_id='order-1', got %s", resp.OrderID)
	}
	if resp.Status != "unassigned" {
		t.Fatalf("expected status='unassigned', got %s", resp.Status)
	}
}

func TestUnassignCourier_InvalidJSON(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockdeliveryService(ctrl)
	h := deliveryHandler.NewDeliveryHandler(mockService)
	r := chi.NewRouter()
	r.Post("/delivery/unassign", h.Unassign)

	req := httptest.NewRequest("POST", "/delivery/unassign", bytes.NewBuffer([]byte("{invalid")))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestUnassignCourier_EmptyOrderID(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockdeliveryService(ctrl)
	h := deliveryHandler.NewDeliveryHandler(mockService)
	r := chi.NewRouter()
	r.Post("/delivery/unassign", h.Unassign)

	body := `{"order_id":""}`
	req := httptest.NewRequest("POST", "/delivery/unassign", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestUnassignCourier_DeliveryNotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockdeliveryService(ctrl)
	mockService.EXPECT().
		UnassignCourier(gomock.Any(), "order-1").
		Return(nil, modelDelivery.ErrDeliveryNotFound)

	h := deliveryHandler.NewDeliveryHandler(mockService)
	r := chi.NewRouter()
	r.Post("/delivery/unassign", h.Unassign)

	body := `{"order_id":"order-1"}`
	req := httptest.NewRequest("POST", "/delivery/unassign", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", rr.Code)
	}
}
