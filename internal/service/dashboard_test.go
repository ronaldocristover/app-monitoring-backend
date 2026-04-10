package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
)

func TestNewDashboardService(t *testing.T) {
	appRepo := new(MockAppRepository)
	svcRepo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	envRepo := new(MockEnvironmentRepository)
	logger := zap.NewNop().Sugar()

	svc := NewDashboardService(appRepo, svcRepo, logRepo, envRepo, nil, logger)

	assert.NotNil(t, svc)
}

func TestDashboardGetDashboardStats_NilDB(t *testing.T) {
	appRepo := new(MockAppRepository)
	svcRepo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	envRepo := new(MockEnvironmentRepository)
	logger := zap.NewNop().Sugar()

	svc := NewDashboardService(appRepo, svcRepo, logRepo, envRepo, nil, logger)

	// GetDashboardStats with nil db should panic or return error
	// Since the service directly uses db.WithContext, nil db will cause a panic
	assert.Panics(t, func() {
		_, _ = svc.GetDashboardStats(context.Background())
	})
}

func TestDashboardService_Interface(t *testing.T) {
	appRepo := new(MockAppRepository)
	svcRepo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	envRepo := new(MockEnvironmentRepository)
	logger := zap.NewNop().Sugar()

	// Verify the service implements DashboardService interface
	var _ DashboardService = NewDashboardService(appRepo, svcRepo, logRepo, envRepo, nil, logger)
}

// TestDashboardStats_Struct verifies DashboardStats struct field defaults
func TestDashboardStats_StructDefaults(t *testing.T) {
	stats := &DashboardStats{}
	assert.Equal(t, int64(0), stats.TotalApps)
	assert.Equal(t, int64(0), stats.TotalServices)
	assert.Equal(t, int64(0), stats.ServicesUp)
	assert.Equal(t, int64(0), stats.ServicesDown)
	assert.Nil(t, stats.RecentIncidents)
	assert.Nil(t, stats.EnvironmentBreakdown)
}

// TestRecentIncident_Fields verifies RecentIncident struct
func TestRecentIncident_Fields(t *testing.T) {
	id := uuid.New()
	incident := RecentIncident{
		ServiceID:    id,
		ServiceName:  "test-service",
		Status:       "down",
		ErrorMessage: "connection refused",
	}
	assert.Equal(t, id, incident.ServiceID)
	assert.Equal(t, "test-service", incident.ServiceName)
	assert.Equal(t, "down", incident.Status)
	assert.Equal(t, "connection refused", incident.ErrorMessage)
}

// TestEnvironmentBreakdownItem_Fields verifies EnvironmentBreakdownItem struct
func TestEnvironmentBreakdownItem_Fields(t *testing.T) {
	item := EnvironmentBreakdownItem{
		AppName:       "my-app",
		EnvironmentID: uuid.New().String(),
		Environment:   "production",
		TotalServices: 5,
	}
	assert.Equal(t, "my-app", item.AppName)
	assert.Equal(t, "production", item.Environment)
	assert.Equal(t, int64(5), item.TotalServices)
}

// TestDashboardStats_WithMonitoringLogModel verifies MonitoringLog model compatibility
func TestDashboardStats_WithMonitoringLogModel(t *testing.T) {
	log := &model.MonitoringLog{
		ID:           uuid.New(),
		ServiceID:    uuid.New(),
		Status:       "down",
		ErrorMessage: "timeout",
	}
	assert.Equal(t, "down", log.Status)
	assert.Equal(t, "timeout", log.ErrorMessage)
}
