package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ronaldocristover/app-monitoring/internal/model"
	"gorm.io/gorm"
)

func createTestServiceForConfig(t *testing.T, db *gorm.DB) *model.Service {
	t.Helper()
	app := &model.App{AppName: "MCApp_" + t.Name()}
	require.NoError(t, db.Create(app).Error)
	server := &model.Server{Name: "MCServer_" + t.Name(), IP: "10.0.0.1"}
	require.NoError(t, db.Create(server).Error)
	env := &model.Environment{AppID: app.ID, Name: "production_" + t.Name()}
	require.NoError(t, db.Create(env).Error)
	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          "MCService_" + t.Name(),
	}
	require.NoError(t, db.Create(svc).Error)
	return svc
}

func createTestServiceWithSuffix(t *testing.T, db *gorm.DB, suffix string) *model.Service {
	t.Helper()
	app := &model.App{AppName: fmt.Sprintf("MCApp_%s_%s", t.Name(), suffix)}
	require.NoError(t, db.Create(app).Error)
	server := &model.Server{Name: fmt.Sprintf("MCServer_%s_%s", t.Name(), suffix), IP: "10.0.0.1"}
	require.NoError(t, db.Create(server).Error)
	env := &model.Environment{AppID: app.ID, Name: fmt.Sprintf("production_%s_%s", t.Name(), suffix)}
	require.NoError(t, db.Create(env).Error)
	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          fmt.Sprintf("MCService_%s_%s", t.Name(), suffix),
	}
	require.NoError(t, db.Create(svc).Error)
	return svc
}

func TestMonitoringConfigRepository_GetByService(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringConfigRepository(db)
	ctx := context.Background()

	svc := createTestServiceForConfig(t, db)

	config := &model.MonitoringConfig{
		ServiceID:           svc.ID,
		Enabled:             true,
		PingIntervalSeconds: 60,
		TimeoutSeconds:      10,
		Retries:             3,
	}
	require.NoError(t, repo.Upsert(ctx, config))

	found, err := repo.GetByService(ctx, svc.ID)
	assert.NoError(t, err)
	assert.Equal(t, svc.ID, found.ServiceID)
	assert.True(t, found.Enabled)
	assert.Equal(t, 60, found.PingIntervalSeconds)
	assert.Equal(t, 10, found.TimeoutSeconds)
	assert.Equal(t, 3, found.Retries)

	// Test non-existent service returns error
	_, err = repo.GetByService(ctx, uuid.New())
	assert.Error(t, err)
}

func TestMonitoringConfigRepository_Upsert_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringConfigRepository(db)
	ctx := context.Background()

	svc := createTestServiceForConfig(t, db)

	config := &model.MonitoringConfig{
		ServiceID:           svc.ID,
		Enabled:             true,
		PingIntervalSeconds: 60,
		TimeoutSeconds:      10,
		Retries:             3,
	}
	err := repo.Upsert(ctx, config)
	assert.NoError(t, err)

	found, err := repo.GetByService(ctx, svc.ID)
	assert.NoError(t, err)
	assert.Equal(t, svc.ID, found.ServiceID)
	assert.True(t, found.Enabled)
	assert.Equal(t, 60, found.PingIntervalSeconds)
	assert.Equal(t, 10, found.TimeoutSeconds)
	assert.Equal(t, 3, found.Retries)
}

func TestMonitoringConfigRepository_Upsert_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringConfigRepository(db)
	ctx := context.Background()

	svc := createTestServiceForConfig(t, db)

	// Initial upsert with enabled=true, PingIntervalSeconds=60
	config := &model.MonitoringConfig{
		ServiceID:           svc.ID,
		Enabled:             true,
		PingIntervalSeconds: 60,
		TimeoutSeconds:      10,
		Retries:             3,
	}
	require.NoError(t, repo.Upsert(ctx, config))

	// Upsert again with different non-zero values
	config2 := &model.MonitoringConfig{
		ServiceID:           svc.ID,
		Enabled:             true,
		PingIntervalSeconds: 30,
		TimeoutSeconds:      5,
		Retries:             1,
	}
	err := repo.Upsert(ctx, config2)
	assert.NoError(t, err)

	found, err := repo.GetByService(ctx, svc.ID)
	assert.NoError(t, err)
	assert.Equal(t, svc.ID, found.ServiceID)
	assert.True(t, found.Enabled)
	assert.Equal(t, 30, found.PingIntervalSeconds)
	assert.Equal(t, 5, found.TimeoutSeconds)
	assert.Equal(t, 1, found.Retries)
}

func TestMonitoringConfigRepository_FindEnabled(t *testing.T) {
	db := setupTestDB(t)
	repo := NewMonitoringConfigRepository(db)
	ctx := context.Background()

	// Create 2 services: one with an enabled config, one with no config at all
	svc1 := createTestServiceWithSuffix(t, db, "enabled")
	config1 := &model.MonitoringConfig{
		ServiceID:           svc1.ID,
		Enabled:             true,
		PingIntervalSeconds: 60,
		TimeoutSeconds:      10,
		Retries:             3,
	}
	require.NoError(t, repo.Upsert(ctx, config1))

	// Second service has no config
	_ = createTestServiceWithSuffix(t, db, "noconfig")

	var configs []model.MonitoringConfig
	err := repo.FindEnabled(ctx, &configs)
	assert.NoError(t, err)
	assert.Len(t, configs, 1)
	assert.Equal(t, svc1.ID, configs[0].ServiceID)
	assert.True(t, configs[0].Enabled)
}
