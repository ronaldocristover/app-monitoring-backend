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

// createTestService creates a full dependency chain (app -> env, server) and a service.
func createTestService(t *testing.T, db *gorm.DB, svcRepo ServiceRepository, suffix string) (*model.Environment, *model.Server, *model.Service) {
	t.Helper()
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)

	app := &model.App{AppName: fmt.Sprintf("SvcApp_%s", suffix)}
	require.NoError(t, appRepo.Create(context.Background(), app))

	env := &model.Environment{AppID: app.ID, Name: fmt.Sprintf("env_%s", suffix)}
	require.NoError(t, envRepo.Create(context.Background(), env))

	server := &model.Server{Name: fmt.Sprintf("srv_%s", suffix), IP: "10.0.0.1"}
	require.NoError(t, db.Create(server).Error)

	svc := &model.Service{
		EnvironmentID:  env.ID,
		ServerID:       server.ID,
		Name:           fmt.Sprintf("service_%s", suffix),
		Type:           "web",
		URL:            fmt.Sprintf("https://%s.example.com", suffix),
		Repository:     fmt.Sprintf("https://github.com/test/%s", suffix),
		StackLanguage:  "go",
		StackFramework: "gin",
		DBType:         "postgres",
		DBHost:         "db.example.com",
	}
	require.NoError(t, svcRepo.Create(context.Background(), svc))
	return env, server, svc
}

func TestServiceRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	_, _, svc := createTestService(t, db, repo, "create")
	assert.NotEqual(t, uuid.Nil, svc.ID)
}

func TestServiceRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	_, _, svc := createTestService(t, db, repo, "get")

	found, err := repo.GetByID(context.Background(), svc.ID)
	assert.NoError(t, err)
	assert.Equal(t, svc.ID, found.ID)
	assert.Equal(t, svc.Name, found.Name)
	assert.Equal(t, svc.EnvironmentID, found.EnvironmentID)
	assert.Equal(t, svc.ServerID, found.ServerID)
}

func TestServiceRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestServiceRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	_, _, svc := createTestService(t, db, repo, "update")
	svc.Name = "updated-service"
	svc.URL = "https://updated.example.com"

	err := repo.Update(context.Background(), svc)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), svc.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-service", found.Name)
	assert.Equal(t, "https://updated.example.com", found.URL)
}

func TestServiceRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	_, _, svc := createTestService(t, db, repo, "delete")

	err := repo.Delete(context.Background(), svc.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), svc.ID)
	assert.Error(t, err)
}

func TestServiceRepository_GetByIDFull(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	env, server, svc := createTestService(t, db, repo, "full")

	// Create related records
	monitoringConfig := &model.MonitoringConfig{
		ServiceID:           svc.ID,
		Enabled:             true,
		PingIntervalSeconds: 30,
		TimeoutSeconds:      5,
		Retries:             3,
	}
	require.NoError(t, db.Create(monitoringConfig).Error)

	deployment := &model.Deployment{
		ServiceID:     svc.ID,
		Method:        "docker",
		ContainerName: "api-container",
		Port:          8080,
	}
	require.NoError(t, db.Create(deployment).Error)

	found, err := repo.GetByIDFull(context.Background(), svc.ID)
	assert.NoError(t, err)
	assert.Equal(t, svc.ID, found.ID)
	assert.NotNil(t, found.Environment)
	assert.Equal(t, env.ID, found.Environment.ID)
	assert.NotNil(t, found.Server)
	assert.Equal(t, server.ID, found.Server.ID)
	assert.NotNil(t, found.MonitoringConfig)
	assert.Equal(t, monitoringConfig.ID, found.MonitoringConfig.ID)
	assert.Len(t, found.Deployments, 1)
}

func TestServiceRepository_List_FilterByEnvironment(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	env1, _, svc1 := createTestService(t, db, repo, "env1")
	_, _, _ = createTestService(t, db, repo, "env2") // different env

	services, total, err := repo.List(context.Background(), &model.ListServicesRequest{
		Page:          1,
		PageSize:      20,
		EnvironmentID: env1.ID.String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, services, 1)
	assert.Equal(t, svc1.ID, services[0].ID)
}

func TestServiceRepository_List_FilterByServer(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	_, server1, svc1 := createTestService(t, db, repo, "srv1")
	_, _, _ = createTestService(t, db, repo, "srv2") // different server

	services, total, err := repo.List(context.Background(), &model.ListServicesRequest{
		Page:     1,
		PageSize: 20,
		ServerID: server1.ID.String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, services, 1)
	assert.Equal(t, svc1.ID, services[0].ID)
}

func TestServiceRepository_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	for i := 0; i < 5; i++ {
		createTestService(t, db, repo, fmt.Sprintf("page_%d", i))
	}

	services, total, err := repo.List(context.Background(), &model.ListServicesRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, services, 2)
}

func TestServiceRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
}

func TestNewServiceRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)
	assert.NotNil(t, repo)
}
