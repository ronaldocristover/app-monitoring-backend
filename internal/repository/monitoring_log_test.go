package repository

import (
	"context"
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
	app := &model.App{AppName: "LogApp_" + t.Name()}
	require.NoError(t, db.Create(app).Error)
	server := &model.Server{Name: "LogServer_" + t.Name(), IP: "10.0.0.2"}
	require.NoError(t, db.Create(server).Error)
	env := &model.Environment{AppID: app.ID, Name: "production_" + t.Name()}
	require.NoError(t, db.Create(env).Error)
	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          "LogService_" + t.Name(),
	}
	require.NoError(t, db.Create(svc).Error)
	return svc
}

func TestMonitoringLogRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringLogRepository(db)
	ctx := context.Background()

	svc := createTestServiceForLog(t, db)

	log := &model.MonitoringLog{
		ServiceID:      svc.ID,
		Status:         "up",
		ResponseTimeMs: 150,
		StatusCode:     200,
		CheckedAt:      time.Now(),
	}

	err := repo.Create(ctx, log)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, log.ID)
}

func TestMonitoringLogRepository_ListByService(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringLogRepository(db)
	ctx := context.Background()

	svc := createTestServiceForLog(t, db)

	baseTime := time.Now()
	for i := 0; i < 3; i++ {
		log := &model.MonitoringLog{
			ServiceID:      svc.ID,
			Status:         "up",
			ResponseTimeMs: 100 + i,
			StatusCode:     200,
			CheckedAt:      baseTime.Add(time.Duration(i) * time.Minute),
		}
		require.NoError(t, repo.Create(ctx, log))
	}

	logs, total, err := repo.ListByService(ctx, svc.ID, &model.ListMonitoringLogsRequest{
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, logs, 3)
}

func TestMonitoringLogRepository_ListByService_WithStatusFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringLogRepository(db)
	ctx := context.Background()

	svc := createTestServiceForLog(t, db)

	now := time.Now()
	// Create 2 "up" logs
	require.NoError(t, repo.Create(ctx, &model.MonitoringLog{
		ServiceID:      svc.ID,
		Status:         "up",
		ResponseTimeMs: 100,
		StatusCode:     200,
		CheckedAt:      now,
	}))
	require.NoError(t, repo.Create(ctx, &model.MonitoringLog{
		ServiceID:      svc.ID,
		Status:         "up",
		ResponseTimeMs: 120,
		StatusCode:     200,
		CheckedAt:      now.Add(time.Minute),
	}))
	// Create 1 "down" log
	require.NoError(t, repo.Create(ctx, &model.MonitoringLog{
		ServiceID:      svc.ID,
		Status:         "down",
		ResponseTimeMs: 0,
		StatusCode:     500,
		ErrorMessage:   "connection refused",
		CheckedAt:      now.Add(2 * time.Minute),
	}))

	logs, total, err := repo.ListByService(ctx, svc.ID, &model.ListMonitoringLogsRequest{
		Page:     1,
		PageSize: 10,
		Status:   "up",
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, logs, 2)
	for _, l := range logs {
		assert.Equal(t, "up", l.Status)
	}
}

func TestMonitoringLogRepository_GetLatest(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringLogRepository(db)
	ctx := context.Background()

	svc := createTestServiceForLog(t, db)

	now := time.Now()
	// Create 3 logs with different CheckedAt times
	require.NoError(t, repo.Create(ctx, &model.MonitoringLog{
		ServiceID:      svc.ID,
		Status:         "up",
		ResponseTimeMs: 100,
		StatusCode:     200,
		CheckedAt:      now.Add(-2 * time.Hour),
	}))
	require.NoError(t, repo.Create(ctx, &model.MonitoringLog{
		ServiceID:      svc.ID,
		Status:         "up",
		ResponseTimeMs: 200,
		StatusCode:     200,
		CheckedAt:      now.Add(-1 * time.Hour),
	}))
	require.NoError(t, repo.Create(ctx, &model.MonitoringLog{
		ServiceID:      svc.ID,
		Status:         "down",
		ResponseTimeMs: 0,
		StatusCode:     500,
		ErrorMessage:   "timeout",
		CheckedAt:      now,
	}))

	latest, err := repo.GetLatest(ctx, svc.ID)
	assert.NoError(t, err)
	assert.Equal(t, "down", latest.Status)
	assert.Equal(t, 0, latest.ResponseTimeMs)
	assert.Equal(t, 500, latest.StatusCode)
	// The most recent log should have CheckedAt close to now
	assert.True(t, latest.CheckedAt.After(now.Add(-1*time.Second)))
}
