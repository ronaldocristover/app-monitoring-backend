package handler

import (
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
	"github.com/ronaldocristover/app-monitoring/pkg/response"
)

func TestMonitoringLogListByService_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringLogService)
	h := NewMonitoringLogHandler(svc)

	serviceID := uuid.New()
	logs := []*model.MonitoringLog{
		{ID: uuid.New(), ServiceID: serviceID, Status: "up"},
		{ID: uuid.New(), ServiceID: serviceID, Status: "down"},
	}

	svc.On("ListByService", mock.Anything, serviceID, mock.AnythingOfType("*model.ListMonitoringLogsRequest")).
		Return(logs, int64(2), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/logs?page=1&page_size=20", nil)
	c.AddParam("id", serviceID.String())

	h.ListByService(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Meta)
	assert.Equal(t, 1, resp.Meta.Page)
	assert.Equal(t, 20, resp.Meta.PageSize)
	assert.Equal(t, int64(2), resp.Meta.TotalItems)

	svc.AssertExpectations(t)
}

func TestMonitoringLogListByService_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringLogService)
	h := NewMonitoringLogHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/invalid/logs", nil)
	c.AddParam("id", "invalid")

	h.ListByService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitoringLogListByService_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringLogService)
	h := NewMonitoringLogHandler(svc)

	serviceID := uuid.New()
	logs := []*model.MonitoringLog{
		{ID: uuid.New(), ServiceID: serviceID, Status: "up"},
	}

	svc.On("ListByService", mock.Anything, serviceID, mock.MatchedBy(func(req *model.ListMonitoringLogsRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(logs, int64(1), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/logs", nil)
	c.AddParam("id", serviceID.String())

	h.ListByService(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Meta.Page)
	assert.Equal(t, 20, resp.Meta.PageSize)

	svc.AssertExpectations(t)
}

func TestMonitoringLogListByService_ServiceNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringLogService)
	h := NewMonitoringLogHandler(svc)

	serviceID := uuid.New()
	svc.On("ListByService", mock.Anything, serviceID, mock.AnythingOfType("*model.ListMonitoringLogsRequest")).
		Return([]*model.MonitoringLog(nil), int64(0), service.ErrServiceNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/logs", nil)
	c.AddParam("id", serviceID.String())

	h.ListByService(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestMonitoringLogListByService_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.MonitoringLogService)
	h := NewMonitoringLogHandler(svc)

	serviceID := uuid.New()
	svc.On("ListByService", mock.Anything, serviceID, mock.AnythingOfType("*model.ListMonitoringLogsRequest")).
		Return([]*model.MonitoringLog(nil), int64(0), errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/logs", nil)
	c.AddParam("id", serviceID.String())

	h.ListByService(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}
