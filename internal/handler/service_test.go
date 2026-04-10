package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ronaldocristover/app-monitoring/internal/handler/mocks"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/internal/service"
	"github.com/ronaldocristover/app-monitoring/pkg/apierror"
)

// --- Service Handler Tests ---

func TestServiceCreate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	envID := uuid.New()
	srvID := uuid.New()
	svcID := uuid.New()
	svc := &model.Service{ID: svcID, Name: "svc1"}
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateServiceRequest")).Return(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := fmt.Sprintf(`{"environment_id":"%s","server_id":"%s","name":"svc1","url":"http://example.com"}`, envID, srvID)
	req, _ := http.NewRequest(http.MethodPost, "/services", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceCreate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, "/services", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceCreate_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	envID := uuid.New()
	srvID := uuid.New()
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateServiceRequest")).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := fmt.Sprintf(`{"environment_id":"%s","server_id":"%s","name":"svc1","url":"http://example.com"}`, envID, srvID)
	req, _ := http.NewRequest(http.MethodPost, "/services", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Create(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceGet_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()
	svc := &model.Service{ID: id, Name: "svc1"}
	mockSvc.On("GetByIDFull", mock.Anything, id).Return(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/services/%s", id), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceGet_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/services/not-a-uuid", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Get(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceGet_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()
	mockSvc.On("GetByIDFull", mock.Anything, id).Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/services/%s", id), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Get(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceUpdate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()
	svc := &model.Service{ID: id, Name: "svc-updated"}
	mockSvc.On("Update", mock.Anything, id, mock.AnythingOfType("*model.UpdateServiceRequest")).Return(svc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"name":"svc-updated"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/services/%s", id), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceUpdate_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"name":"svc-updated"}`
	req, _ := http.NewRequest(http.MethodPut, "/services/not-a-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceUpdate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/services/%s", id), bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceUpdate_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()
	mockSvc.On("Update", mock.Anything, id, mock.AnythingOfType("*model.UpdateServiceRequest")).Return(nil, service.ErrServiceNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"name":"svc-updated"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/services/%s", id), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceUpdate_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()
	mockSvc.On("Update", mock.Anything, id, mock.AnythingOfType("*model.UpdateServiceRequest")).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"name":"svc-updated"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/services/%s", id), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceDelete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()
	mockSvc.On("Delete", mock.Anything, id).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/services/%s", id), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	mockSvc.AssertExpectations(t)
}

func TestServiceDelete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodDelete, "/services/not-a-uuid", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceDelete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()
	mockSvc.On("Delete", mock.Anything, id).Return(service.ErrServiceNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/services/%s", id), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceDelete_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()
	mockSvc.On("Delete", mock.Anything, id).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/services/%s", id), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceList_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	services := []*model.Service{{ID: uuid.New(), Name: "svc1"}}
	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListServicesRequest")).Return(services, int64(1), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/services?page=1&page_size=20", nil)
	c.Request = req

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["meta"])
	mockSvc.AssertExpectations(t)
}

func TestServiceList_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	services := []*model.Service{{ID: uuid.New(), Name: "svc1"}}
	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListServicesRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(services, int64(1), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/services", nil)
	c.Request = req

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceList_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	mockSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListServicesRequest")).Return(nil, int64(0), errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/services?page=1&page_size=20", nil)
	c.Request = req

	h.List(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceManualPing_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()
	log := &model.MonitoringLog{ID: uuid.New(), ServiceID: id, Status: "up"}
	mockSvc.On("ManualPing", mock.Anything, id).Return(log, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/services/%s/ping", id), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.ManualPing(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceManualPing_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, "/services/not-a-uuid/ping", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.ManualPing(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceManualPing_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()
	mockSvc.On("ManualPing", mock.Anything, id).Return(nil, service.ErrServiceNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/services/%s/ping", id), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.ManualPing(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceManualPing_APIError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()
	mockSvc.On("ManualPing", mock.Anything, id).Return(nil, apierror.BadRequest("Service has no URL configured"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/services/%s/ping", id), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.ManualPing(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestServiceManualPing_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := new(mocks.ServiceService)
	h := NewServiceHandler(mockSvc)

	id := uuid.New()
	mockSvc.On("ManualPing", mock.Anything, id).Return(nil, errors.New("network timeout"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/services/%s/ping", id), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.ManualPing(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}
