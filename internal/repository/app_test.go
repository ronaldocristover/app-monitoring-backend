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

func createTestApp(t *testing.T, repo AppRepository, suffix string) *model.App {
	t.Helper()
	app := &model.App{
		AppName:     fmt.Sprintf("TestApp_%s", suffix),
		Description: "A test application",
		Tags:        "backend,api",
	}
	require.NoError(t, repo.Create(context.Background(), app))
	return app
}

func TestAppRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppRepository(db)

	app := &model.App{
		AppName:     "MyApp",
		Description: "Test description",
		Tags:        "frontend,react",
	}

	err := repo.Create(context.Background(), app)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, app.ID)
}

func TestAppRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppRepository(db)

	app := createTestApp(t, repo, "get")

	found, err := repo.GetByID(context.Background(), app.ID)
	assert.NoError(t, err)
	assert.Equal(t, app.ID, found.ID)
	assert.Equal(t, app.AppName, found.AppName)
	assert.Equal(t, app.Description, found.Description)
}

func TestAppRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestAppRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppRepository(db)

	app := createTestApp(t, repo, "update")
	app.AppName = "UpdatedApp"
	app.Description = "Updated description"

	err := repo.Update(context.Background(), app)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), app.ID)
	assert.NoError(t, err)
	assert.Equal(t, "UpdatedApp", found.AppName)
	assert.Equal(t, "Updated description", found.Description)
}

func TestAppRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppRepository(db)

	app := createTestApp(t, repo, "delete")

	err := repo.Delete(context.Background(), app.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), app.ID)
	assert.Error(t, err)
}

func TestAppRepository_Delete_CascadesEnvironmentsAndServices(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)
	svcRepo := NewServiceRepository(db)

	app := createTestApp(t, appRepo, "cascade")

	// Create a server for the service
	server := &model.Server{Name: "srv", IP: "10.0.0.1"}
	require.NoError(t, db.Create(server).Error)

	env := &model.Environment{AppID: app.ID, Name: "prod"}
	require.NoError(t, envRepo.Create(context.Background(), env))

	svc := &model.Service{EnvironmentID: env.ID, ServerID: server.ID, Name: "api"}
	require.NoError(t, svcRepo.Create(context.Background(), svc))

	// Delete app should cascade
	err := appRepo.Delete(context.Background(), app.ID)
	assert.NoError(t, err)

	_, err = envRepo.GetByID(context.Background(), env.ID)
	assert.Error(t, err)

	_, err = svcRepo.GetByID(context.Background(), svc.ID)
	assert.Error(t, err)
}

func TestAppRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppRepository(db)

	for i := 0; i < 5; i++ {
		createTestApp(t, repo, fmt.Sprintf("list_%d", i))
	}

	apps, total, err := repo.List(context.Background(), &model.ListAppsRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, apps, 5)
}

func TestAppRepository_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppRepository(db)

	for i := 0; i < 5; i++ {
		createTestApp(t, repo, fmt.Sprintf("page_%d", i))
	}

	apps, total, err := repo.List(context.Background(), &model.ListAppsRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, apps, 2)
}

func TestAppRepository_List_WithTagFilter(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
}

func TestAppRepository_GetByIDFull(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)
	svcRepo := NewServiceRepository(db)

	app := createTestApp(t, appRepo, "full")

	server := &model.Server{Name: "srv", IP: "10.0.0.1"}
	require.NoError(t, db.Create(server).Error)

	env := &model.Environment{AppID: app.ID, Name: "prod"}
	require.NoError(t, envRepo.Create(context.Background(), env))

	svc := &model.Service{EnvironmentID: env.ID, ServerID: server.ID, Name: "api"}
	require.NoError(t, svcRepo.Create(context.Background(), svc))

	found, err := appRepo.GetByIDFull(context.Background(), app.ID)
	assert.NoError(t, err)
	assert.Equal(t, app.ID, found.ID)
	assert.Len(t, found.Environments, 1)
	assert.Equal(t, env.ID, found.Environments[0].ID)
	assert.Len(t, found.Environments[0].Services, 1)
	assert.Equal(t, svc.ID, found.Environments[0].Services[0].ID)
}

func TestAppRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
}

func TestNewAppRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppRepository(db)
	assert.NotNil(t, repo)
}
