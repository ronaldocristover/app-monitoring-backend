package handler

import (
	"bytes"
	"encoding/json"
	"errors"
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
)

func TestMonitoringConfigGet_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringConfigService)
	h := NewMonitoringConfigHandler(svc)

	serviceID := uuid.New()
	config := &model.MonitoringConfig{
		ID:        uuid.New(),
		ServiceID: serviceID,
		Enabled:   true,
	}
	svc.On("GetByService", mock.Anything, serviceID).Return(config, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/monitoring", nil)
	c.AddParam("id", serviceID.String())

	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestMonitoringConfigGet_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringConfigService)
	h := NewMonitoringConfigHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/invalid/monitoring", nil)
	c.AddParam("id", "invalid")

	h.Get(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitoringConfigGet_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringConfigService)
	h := NewMonitoringConfigHandler(svc)

	serviceID := uuid.New()
	svc.On("GetByService", mock.Anything, serviceID).Return(nil, service.ErrMonitoringConfigNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/monitoring", nil)
	c.AddParam("id", serviceID.String())

	h.Get(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestMonitoringConfigGet_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringConfigService)
	h := NewMonitoringConfigHandler(svc)

	serviceID := uuid.New()
	svc.On("GetByService", mock.Anything, serviceID).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/monitoring", nil)
	c.AddParam("id", serviceID.String())

	h.Get(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

func TestMonitoringConfigUpsert_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringConfigService)
	h := NewMonitoringConfigHandler(svc)

	serviceID := uuid.New()
	config := &model.MonitoringConfig{
		ID:        uuid.New(),
		ServiceID: serviceID,
		Enabled:   true,
	}

	svc.On("Upsert", mock.Anything, serviceID, mock.AnythingOfType("*model.UpdateMonitoringConfigRequest")).Return(config, nil)

	enabled := true
	body := model.UpdateMonitoringConfigRequest{
		Enabled:             &enabled,
		PingIntervalSeconds: intPtr(30),
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/services/"+serviceID.String()+"/monitoring", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", serviceID.String())

	h.Upsert(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestMonitoringConfigUpsert_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringConfigService)
	h := NewMonitoringConfigHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/services/invalid/monitoring", bytes.NewReader([]byte(`{}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", "invalid")

	h.Upsert(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitoringConfigUpsert_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringConfigService)
	h := NewMonitoringConfigHandler(svc)

	serviceID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/services/"+serviceID.String()+"/monitoring", bytes.NewReader([]byte(`invalid json`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", serviceID.String())

	h.Upsert(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitoringConfigUpsert_ServiceNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringConfigService)
	h := NewMonitoringConfigHandler(svc)

	serviceID := uuid.New()
	svc.On("Upsert", mock.Anything, serviceID, mock.AnythingOfType("*model.UpdateMonitoringConfigRequest")).Return(nil, service.ErrServiceNotFound)

	enabled := true
	body := model.UpdateMonitoringConfigRequest{Enabled: &enabled}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/services/"+serviceID.String()+"/monitoring", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", serviceID.String())

	h.Upsert(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestMonitoringConfigUpsert_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringConfigService)
	h := NewMonitoringConfigHandler(svc)

	serviceID := uuid.New()
	svc.On("Upsert", mock.Anything, serviceID, mock.AnythingOfType("*model.UpdateMonitoringConfigRequest")).Return(nil, errors.New("db error"))

	enabled := true
	body := model.UpdateMonitoringConfigRequest{Enabled: &enabled}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/services/"+serviceID.String()+"/monitoring", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", serviceID.String())

	h.Upsert(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// Helper to create int pointer
func intPtr(v int) *int {
	return &v
}
