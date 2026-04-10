package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ronaldocristover/app-monitoring/internal/handler/mocks"
	"github.com/ronaldocristover/app-monitoring/internal/service"
)

func TestDashboardGetStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DashboardService)
	h := NewDashboardHandler(svc)

	expected := &service.DashboardStats{
		TotalApps:     5,
		TotalServices: 12,
		ServicesUp:    10,
		ServicesDown:  2,
		RecentIncidents: []service.RecentIncident{
			{ServiceName: "api-gateway", Status: "down"},
		},
		EnvironmentBreakdown: []service.EnvironmentBreakdownItem{
			{AppName: "my-app", Environment: "production", TotalServices: 3},
		},
	}

	svc.On("GetDashboardStats", mock.Anything).Return(expected, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)

	h.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(5), data["total_apps"])
	assert.Equal(t, float64(12), data["total_services"])
	assert.Equal(t, float64(10), data["services_up"])
	assert.Equal(t, float64(2), data["services_down"])

	incidents := data["recent_incidents"].([]interface{})
	assert.Len(t, incidents, 1)

	breakdown := data["environment_breakdown"].([]interface{})
	assert.Len(t, breakdown, 1)

	svc.AssertExpectations(t)
}

func TestDashboardGetStats_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DashboardService)
	h := NewDashboardHandler(svc)

	expected := &service.DashboardStats{
		TotalApps:            0,
		TotalServices:        0,
		ServicesUp:           0,
		ServicesDown:         0,
		RecentIncidents:      []service.RecentIncident{},
		EnvironmentBreakdown: []service.EnvironmentBreakdownItem{},
	}

	svc.On("GetDashboardStats", mock.Anything).Return(expected, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)

	h.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(0), data["total_apps"])
	assert.Equal(t, float64(0), data["total_services"])

	svc.AssertExpectations(t)
}

func TestDashboardGetStats_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DashboardService)
	h := NewDashboardHandler(svc)

	svc.On("GetDashboardStats", mock.Anything).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)

	h.GetStats(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["success"].(bool))

	svc.AssertExpectations(t)
}
