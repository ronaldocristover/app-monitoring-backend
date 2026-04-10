package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ronaldocristover/app-monitoring/internal/model"
)

func TestEnvironmentRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)

	app := &model.App{AppName: "App1"}
	require.NoError(t, appRepo.Create(context.Background(), app))

	env := &model.Environment{
		AppID: app.ID,
		Name:  "production",
	}

	err := envRepo.Create(context.Background(), env)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, env.ID)
}

func TestEnvironmentRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)

	app := &model.App{AppName: "App1"}
	require.NoError(t, appRepo.Create(context.Background(), app))

	env := &model.Environment{AppID: app.ID, Name: "staging"}
	require.NoError(t, envRepo.Create(context.Background(), env))

	found, err := envRepo.GetByID(context.Background(), env.ID)
	assert.NoError(t, err)
	assert.Equal(t, env.ID, found.ID)
	assert.Equal(t, env.Name, found.Name)
	assert.Equal(t, app.ID, found.AppID)
}

func TestEnvironmentRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	envRepo := NewEnvironmentRepository(db)

	_, err := envRepo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestEnvironmentRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)

	app := &model.App{AppName: "App1"}
	require.NoError(t, appRepo.Create(context.Background(), app))

	env := &model.Environment{AppID: app.ID, Name: "staging"}
	require.NoError(t, envRepo.Create(context.Background(), env))

	env.Name = "production"
	err := envRepo.Update(context.Background(), env)
	assert.NoError(t, err)

	found, err := envRepo.GetByID(context.Background(), env.ID)
	assert.NoError(t, err)
	assert.Equal(t, "production", found.Name)
}

func TestEnvironmentRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)

	app := &model.App{AppName: "App1"}
	require.NoError(t, appRepo.Create(context.Background(), app))

	env := &model.Environment{AppID: app.ID, Name: "staging"}
	require.NoError(t, envRepo.Create(context.Background(), env))

	err := envRepo.Delete(context.Background(), env.ID)
	assert.NoError(t, err)

	_, err = envRepo.GetByID(context.Background(), env.ID)
	assert.Error(t, err)
}

func TestEnvironmentRepository_Delete_CascadesServices(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)
	svcRepo := NewServiceRepository(db)

	app := &model.App{AppName: "App1"}
	require.NoError(t, appRepo.Create(context.Background(), app))

	server := &model.Server{Name: "srv", IP: "10.0.0.1"}
	require.NoError(t, db.Create(server).Error)

	env := &model.Environment{AppID: app.ID, Name: "prod"}
	require.NoError(t, envRepo.Create(context.Background(), env))

	svc := &model.Service{EnvironmentID: env.ID, ServerID: server.ID, Name: "api"}
	require.NoError(t, svcRepo.Create(context.Background(), svc))

	err := envRepo.Delete(context.Background(), env.ID)
	assert.NoError(t, err)

	_, err = svcRepo.GetByID(context.Background(), svc.ID)
	assert.Error(t, err)
}

func TestEnvironmentRepository_List(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)

	app := &model.App{AppName: "App1"}
	require.NoError(t, appRepo.Create(context.Background(), app))

	for i := 0; i < 3; i++ {
		env := &model.Environment{
			AppID: app.ID,
			Name:  fmt.Sprintf("env_%d", i),
		}
		require.NoError(t, envRepo.Create(context.Background(), env))
	}

	envs, total, err := envRepo.List(context.Background(), &model.ListEnvironmentsRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, envs, 3)
}

func TestEnvironmentRepository_List_FilterByAppID(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)

	app1 := &model.App{AppName: "App1"}
	app2 := &model.App{AppName: "App2"}
	require.NoError(t, appRepo.Create(context.Background(), app1))
	require.NoError(t, appRepo.Create(context.Background(), app2))

	env1 := &model.Environment{AppID: app1.ID, Name: "prod"}
	env2 := &model.Environment{AppID: app2.ID, Name: "staging"}
	require.NoError(t, envRepo.Create(context.Background(), env1))
	require.NoError(t, envRepo.Create(context.Background(), env2))

	envs, total, err := envRepo.List(context.Background(), &model.ListEnvironmentsRequest{
		Page:     1,
		PageSize: 20,
		AppID:    app1.ID.String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, envs, 1)
	assert.Equal(t, env1.ID, envs[0].ID)
}

func TestEnvironmentRepository_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)

	app := &model.App{AppName: "App1"}
	require.NoError(t, appRepo.Create(context.Background(), app))

	for i := 0; i < 5; i++ {
		env := &model.Environment{AppID: app.ID, Name: fmt.Sprintf("env_%d", i)}
		require.NoError(t, envRepo.Create(context.Background(), env))
	}

	envs, total, err := envRepo.List(context.Background(), &model.ListEnvironmentsRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, envs, 2)
}

func TestNewEnvironmentRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewEnvironmentRepository(db)
	assert.NotNil(t, repo)
}
