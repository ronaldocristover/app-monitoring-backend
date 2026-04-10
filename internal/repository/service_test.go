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

// createTestServiceDependencies creates App, Server, Environment and returns them.
func createTestServiceDependencies(t *testing.T, db *gorm.DB) (*model.App, *model.Server, *model.Environment) {
	t.Helper()
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)

	suffix := uuid.New().String()[:8]

	app := &model.App{AppName: fmt.Sprintf("SvcDepApp_%s", suffix)}
	require.NoError(t, appRepo.Create(context.Background(), app))

	server := &model.Server{
		Name: fmt.Sprintf("srv_dep_%s", suffix),
		IP:   fmt.Sprintf("10.0.%s.1", suffix[:3]),
	}
	require.NoError(t, db.Create(server).Error)

	env := &model.Environment{
		AppID: app.ID,
		Name:  fmt.Sprintf("env_dep_%s", suffix),
	}
	require.NoError(t, envRepo.Create(context.Background(), env))

	return app, server, env
}

func TestServiceRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)
	_, server, env := createTestServiceDependencies(t, db)

	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          fmt.Sprintf("create_svc_%s", uuid.New().String()[:8]),
	}

	err := repo.Create(context.Background(), svc)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, svc.ID)
}

func TestServiceRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)
	_, server, env := createTestServiceDependencies(t, db)

	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          fmt.Sprintf("get_svc_%s", uuid.New().String()[:8]),
	}
	require.NoError(t, repo.Create(context.Background(), svc))

	found, err := repo.GetByID(context.Background(), svc.ID)
	assert.NoError(t, err)
	assert.Equal(t, svc.ID, found.ID)
	assert.Equal(t, svc.Name, found.Name)
	assert.Equal(t, svc.EnvironmentID, found.EnvironmentID)
	assert.Equal(t, svc.ServerID, found.ServerID)

	// Test non-existent ID
	_, err = repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestServiceRepository_GetByIDFull(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)
	_, server, env := createTestServiceDependencies(t, db)

	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          fmt.Sprintf("full_svc_%s", uuid.New().String()[:8]),
	}
	require.NoError(t, repo.Create(context.Background(), svc))

	// Create MonitoringConfig for the service
	monitoringConfig := &model.MonitoringConfig{
		ServiceID:           svc.ID,
		Enabled:             true,
		PingIntervalSeconds: 30,
		TimeoutSeconds:      5,
		Retries:             3,
	}
	require.NoError(t, db.Create(monitoringConfig).Error)

	// Create Backup for the service
	backup := &model.Backup{
		ServiceID: svc.ID,
		Enabled:   true,
		Path:      "/backups/full",
		Schedule:  "daily",
		Status:    "active",
	}
	require.NoError(t, db.Create(backup).Error)

	// Create Deployment for the service
	deployment := &model.Deployment{
		ServiceID:     svc.ID,
		Method:        "docker",
		ContainerName: "full-container",
		Port:          8080,
	}
	require.NoError(t, db.Create(deployment).Error)

	found, err := repo.GetByIDFull(context.Background(), svc.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found.Environment)
	assert.NotNil(t, found.Server)
	assert.NotNil(t, found.MonitoringConfig)
	assert.Len(t, found.Backups, 1)
	assert.Len(t, found.Deployments, 1)
}

func TestServiceRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)
	_, server, env := createTestServiceDependencies(t, db)

	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          fmt.Sprintf("update_svc_%s", uuid.New().String()[:8]),
	}
	require.NoError(t, repo.Create(context.Background(), svc))

	svc.Name = "updated-service-name"
	err := repo.Update(context.Background(), svc)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), svc.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-service-name", found.Name)
}

func TestServiceRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)
	_, server, env := createTestServiceDependencies(t, db)

	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          fmt.Sprintf("delete_svc_%s", uuid.New().String()[:8]),
	}
	require.NoError(t, repo.Create(context.Background(), svc))

	err := repo.Delete(context.Background(), svc.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), svc.ID)
	assert.Error(t, err)
}

func TestServiceRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)
	_, server, env := createTestServiceDependencies(t, db)

	// Count before inserting
	_, totalBefore, err := repo.List(context.Background(), &model.ListServicesRequest{Page: 1, PageSize: 100})
	require.NoError(t, err)

	// Create 3 services (same env+server)
	for i := 0; i < 3; i++ {
		svc := &model.Service{
			EnvironmentID: env.ID,
			ServerID:      server.ID,
			Name:          fmt.Sprintf("list_svc_%d_%s", i, uuid.New().String()[:8]),
		}
		require.NoError(t, repo.Create(context.Background(), svc))
	}

	services, total, err := repo.List(context.Background(), &model.ListServicesRequest{
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, totalBefore+3, int(total))
	assert.Len(t, services, 3)
}

func TestServiceRepository_ListByEnvironment(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)
	envRepo := NewEnvironmentRepository(db)

	// Create 2 envs (2 apps)
	app1 := &model.App{AppName: fmt.Sprintf("EnvApp1_%s", uuid.New().String()[:8])}
	require.NoError(t, db.Create(app1).Error)
	env1 := &model.Environment{AppID: app1.ID, Name: "env1"}
	require.NoError(t, envRepo.Create(context.Background(), env1))

	app2 := &model.App{AppName: fmt.Sprintf("EnvApp2_%s", uuid.New().String()[:8])}
	require.NoError(t, db.Create(app2).Error)
	env2 := &model.Environment{AppID: app2.ID, Name: "env2"}
	require.NoError(t, envRepo.Create(context.Background(), env2))

	server := &model.Server{Name: "env-srv", IP: "10.1.1.1"}
	require.NoError(t, db.Create(server).Error)

	// Create 2 services in env1
	for i := 0; i < 2; i++ {
		svc := &model.Service{
			EnvironmentID: env1.ID,
			ServerID:      server.ID,
			Name:          fmt.Sprintf("env1_svc_%d_%s", i, uuid.New().String()[:8]),
		}
		require.NoError(t, repo.Create(context.Background(), svc))
	}

	// Create 1 service in env2
	svc2 := &model.Service{
		EnvironmentID: env2.ID,
		ServerID:      server.ID,
		Name:          fmt.Sprintf("env2_svc_%s", uuid.New().String()[:8]),
	}
	require.NoError(t, repo.Create(context.Background(), svc2))

	filter := &model.ListServicesRequest{
		Page:          1,
		PageSize:      10,
		EnvironmentID: env1.ID.String(),
	}
	services, total, err := repo.List(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, services, 2)
}

func TestServiceRepository_ListByServer(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)
	envRepo := NewEnvironmentRepository(db)

	app := &model.App{AppName: fmt.Sprintf("SrvApp_%s", uuid.New().String()[:8])}
	require.NoError(t, db.Create(app).Error)
	env := &model.Environment{AppID: app.ID, Name: "env"}
	require.NoError(t, envRepo.Create(context.Background(), env))

	// Create 2 servers
	server1 := &model.Server{Name: "srv1", IP: "10.2.1.1"}
	require.NoError(t, db.Create(server1).Error)
	server2 := &model.Server{Name: "srv2", IP: "10.2.1.2"}
	require.NoError(t, db.Create(server2).Error)

	// Create 2 services on server1
	for i := 0; i < 2; i++ {
		svc := &model.Service{
			EnvironmentID: env.ID,
			ServerID:      server1.ID,
			Name:          fmt.Sprintf("srv1_svc_%d_%s", i, uuid.New().String()[:8]),
		}
		require.NoError(t, repo.Create(context.Background(), svc))
	}

	// Create 1 service on server2
	svc2 := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server2.ID,
		Name:          fmt.Sprintf("srv2_svc_%s", uuid.New().String()[:8]),
	}
	require.NoError(t, repo.Create(context.Background(), svc2))

	filter := &model.ListServicesRequest{
		Page:     1,
		PageSize: 10,
		ServerID: server1.ID.String(),
	}
	services, total, err := repo.List(context.Background(), filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, services, 2)
}
