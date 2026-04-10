package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ronaldocristover/app-monitoring/internal/model"
	"gorm.io/gorm"
)

func createTestServiceForLog(t *testing.T, db *gorm.DB) *model.Service {
	t.Helper()
	app := &model.App{AppName: "LogApp"}
	require.NoError(t, db.Create(app).Error)

	env := &model.Environment{AppID: app.ID, Name: "prod"}
	require.NoError(t, db.Create(env).Error)

	server := &model.Server{Name: "srv", IP: "10.0.0.1"}
	require.NoError(t, db.Create(server).Error)

	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          "logged-service",
	}
	require.NoError(t, db.Create(svc).Error)
	return svc
}

func TestMonitoringLogRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringLogRepository(db)

	svc := createTestServiceForLog(t, db)

	log := &model.MonitoringLog{
		ServiceID:      svc.ID,
		Status:         "up",
		ResponseTimeMs: 150,
		StatusCode:     200,
		CheckedAt:      time.Now(),
	}

	err := repo.Create(context.Background(), log)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, log.ID)
}

func TestMonitoringLogRepository_ListByService(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringLogRepository(db)

	svc := createTestServiceForLog(t, db)

	for i := 0; i < 5; i++ {
		log := &model.MonitoringLog{
			ServiceID:      svc.ID,
			Status:         "up",
			ResponseTimeMs: 100 + i,
			StatusCode:     200,
			CheckedAt:      time.Now().Add(time.Duration(i) * time.Minute),
		}
		require.NoError(t, repo.Create(context.Background(), log))
	}

	logs, total, err := repo.ListByService(context.Background(), svc.ID, &model.ListMonitoringLogsRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, logs, 5)
}

func TestMonitoringLogRepository_ListByService_WithStatusFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringLogRepository(db)

	svc := createTestServiceForLog(t, db)

	upLog := &model.MonitoringLog{
		ServiceID:      svc.ID,
		Status:         "up",
		ResponseTimeMs: 100,
		StatusCode:     200,
		CheckedAt:      time.Now(),
	}
	downLog := &model.MonitoringLog{
		ServiceID:      svc.ID,
		Status:         "down",
		ResponseTimeMs: 0,
		StatusCode:     500,
		ErrorMessage:   "connection refused",
		CheckedAt:      time.Now().Add(time.Minute),
	}
	require.NoError(t, repo.Create(context.Background(), upLog))
	require.NoError(t, repo.Create(context.Background(), downLog))

	logs, total, err := repo.ListByService(context.Background(), svc.ID, &model.ListMonitoringLogsRequest{
		Page:     1,
		PageSize: 20,
		Status:   "up",
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, logs, 1)
	assert.Equal(t, "up", logs[0].Status)
}

func TestMonitoringLogRepository_ListByService_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringLogRepository(db)

	svc := createTestServiceForLog(t, db)

	for i := 0; i < 5; i++ {
		log := &model.MonitoringLog{
			ServiceID:      svc.ID,
			Status:         "up",
			ResponseTimeMs: 100,
			StatusCode:     200,
			CheckedAt:      time.Now().Add(time.Duration(i) * time.Minute),
		}
		require.NoError(t, repo.Create(context.Background(), log))
	}

	logs, total, err := repo.ListByService(context.Background(), svc.ID, &model.ListMonitoringLogsRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, logs, 2)
}

func TestMonitoringLogRepository_GetLatest(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringLogRepository(db)

	svc := createTestServiceForLog(t, db)

	// Create logs with different timestamps
	for i := 0; i < 3; i++ {
		log := &model.MonitoringLog{
			ServiceID:      svc.ID,
			Status:         "up",
			ResponseTimeMs: 100 + i,
			StatusCode:     200,
			CheckedAt:      time.Now().Add(time.Duration(i) * time.Hour),
		}
		require.NoError(t, repo.Create(context.Background(), log))
	}

	latest, err := repo.GetLatest(context.Background(), svc.ID)
	assert.NoError(t, err)
	assert.Equal(t, 102, latest.ResponseTimeMs) // the last one created
}

func TestMonitoringLogRepository_GetLatest_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringLogRepository(db)

	_, err := repo.GetLatest(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestMonitoringLogRepository_ListByService_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringLogRepository(db)

	svc := createTestServiceForLog(t, db)

	// Query logs for a service that has none via another service
	otherSvc := &model.Service{
		EnvironmentID: svc.EnvironmentID,
		ServerID:      svc.ServerID,
		Name:          fmt.Sprintf("other-%s", uuid.New().String()[:8]),
	}
	require.NoError(t, db.Create(otherSvc).Error)

	logs, total, err := repo.ListByService(context.Background(), otherSvc.ID, &model.ListMonitoringLogsRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, logs, 0)
}

func TestNewMonitoringLogRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringLogRepository(db)
	assert.NotNil(t, repo)
}
