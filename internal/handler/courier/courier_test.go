package courier_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"

	courierHandler "service-courier/internal/handler/courier"
	"service-courier/internal/handler/courier/mocks"
	model "service-courier/internal/model/courier"
)

func TestGetCourier_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	mockService.EXPECT().
		GetCourier(gomock.Any(), int64(1)).
		Return(&model.Courier{
			ID:     1,
			Name:   "Ivan",
			Phone:  "+78005553535",
			Status: "available",
		}, nil)

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Get("/courier/{id}", h.Get)

	req := httptest.NewRequest("GET", "/courier/1", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
}

func TestGetCourier_InvalidID(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Get("/courier/{id}", h.Get)

	req := httptest.NewRequest("GET", "/courier/abc", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestGetCourier_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	mockService.EXPECT().
		GetCourier(gomock.Any(), int64(5)).
		Return(nil, model.ErrCourierNotFound)

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Get("/courier/{id}", h.Get)

	req := httptest.NewRequest("GET", "/courier/5", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", rr.Code)
	}
}

func TestGetAllCouriers_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	mockService.EXPECT().
		GetAllCouriers(gomock.Any()).
		Return([]model.Courier{
			{ID: 1, Name: "A"},
			{ID: 2, Name: "B"},
		}, nil)

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Get("/couriers", h.GetAll)

	req := httptest.NewRequest("GET", "/couriers", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
}

func TestGetAllCouriers_ServiceError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	mockService.EXPECT().
		GetAllCouriers(gomock.Any()).
		Return(nil, errors.New("db error"))

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Get("/couriers", h.GetAll)

	req := httptest.NewRequest("GET", "/couriers", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 Internal Server Error, got %d", rr.Code)
	}
}

func TestCreateCourier_InvalidJSON(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Post("/courier", h.Create)

	req := httptest.NewRequest("POST", "/courier", bytes.NewBuffer([]byte("{invalid")))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestCreateCourier_InvalidRequest(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Post("/courier", h.Create)

	body := `{"name": ""}`
	req := httptest.NewRequest("POST", "/courier", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestCreateCourier_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	mockService.EXPECT().
		CreateCourier(gomock.Any(), gomock.Any()).
		Return(int64(10), nil)

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Post("/courier", h.Create)

	body := `{"name": "Ivan", "status": "available", "phone": "+78005553535", "transport_type": "car"}`
	req := httptest.NewRequest("POST", "/courier", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", rr.Code)
	}
}

func TestCreateCourier_PhoneExists(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	mockService.EXPECT().
		CreateCourier(gomock.Any(), gomock.Any()).
		Return(int64(0), model.ErrPhoneExists)

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Post("/courier", h.Create)

	body := `{"name": "Ivan", "status": "available", "phone": "+78005553535", "transport_type": "car"}`
	req := httptest.NewRequest("POST", "/courier", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d", rr.Code)
	}
}

func TestUpdateCourier_InvalidJSON(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Put("/courier", h.Update)

	req := httptest.NewRequest("PUT", "/courier", bytes.NewBuffer([]byte("{invalid")))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestUpdateCourier_InvalidRequest(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Put("/courier", h.Update)

	body := `{"id":0}`
	req := httptest.NewRequest("PUT", "/courier", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rr.Code)
	}
}

func TestUpdateCourier_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	mockService.EXPECT().
		UpdateCourier(gomock.Any(), gomock.Any()).
		Return(nil)

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Put("/courier", h.Update)

	body := `{"id": 1, "name": "Updated Name"}`
	req := httptest.NewRequest("PUT", "/courier", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
}

func TestUpdateCourier_NotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockService := mocks.NewMockcourierService(ctrl)

	mockService.EXPECT().
		UpdateCourier(gomock.Any(), gomock.Any()).
		Return(model.ErrCourierNotFound)

	h := courierHandler.NewCourierHandler(mockService)

	r := chi.NewRouter()
	r.Put("/courier", h.Update)

	body := `{"id": 55, "name": "X"}`
	req := httptest.NewRequest("PUT", "/courier", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", rr.Code)
	}
}
