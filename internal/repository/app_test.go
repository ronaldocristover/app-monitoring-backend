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

func TestAppRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppRepository(db)

	app := &model.App{
		AppName:     fmt.Sprintf("CreateApp_%s", uuid.New().String()[:8]),
		Description: "Test create app",
		Tags:        "backend,api",
	}

	err := repo.Create(context.Background(), app)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, app.ID)
}

func TestAppRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppRepository(db)

	app := &model.App{
		AppName:     fmt.Sprintf("GetApp_%s", uuid.New().String()[:8]),
		Description: "Test get app",
	}
	require.NoError(t, repo.Create(context.Background(), app))

	found, err := repo.GetByID(context.Background(), app.ID)
	assert.NoError(t, err)
	assert.Equal(t, app.ID, found.ID)
	assert.Equal(t, app.AppName, found.AppName)
	assert.Equal(t, app.Description, found.Description)

	// Test non-existent ID
	_, err = repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestAppRepository_GetByIDFull(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)
	svcRepo := NewServiceRepository(db)

	app := &model.App{
		AppName: fmt.Sprintf("FullApp_%s", uuid.New().String()[:8]),
	}
	require.NoError(t, appRepo.Create(context.Background(), app))

	server := &model.Server{Name: "full-srv", IP: "10.0.0.1"}
	require.NoError(t, db.Create(server).Error)

	env := &model.Environment{AppID: app.ID, Name: "production"}
	require.NoError(t, envRepo.Create(context.Background(), env))

	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          "full-svc",
	}
	require.NoError(t, svcRepo.Create(context.Background(), svc))

	found, err := appRepo.GetByIDFull(context.Background(), app.ID)
	assert.NoError(t, err)
	assert.Equal(t, app.ID, found.ID)
	assert.Len(t, found.Environments, 1)
	assert.Len(t, found.Environments[0].Services, 1)
}

func TestAppRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppRepository(db)

	app := &model.App{
		AppName:     fmt.Sprintf("UpdateApp_%s", uuid.New().String()[:8]),
		Description: "Original description",
	}
	require.NoError(t, repo.Create(context.Background(), app))

	app.AppName = "UpdatedAppName"
	app.Description = "Updated description"
	err := repo.Update(context.Background(), app)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), app.ID)
	assert.NoError(t, err)
	assert.Equal(t, "UpdatedAppName", found.AppName)
	assert.Equal(t, "Updated description", found.Description)
}

func TestAppRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)
	svcRepo := NewServiceRepository(db)

	app := &model.App{
		AppName: fmt.Sprintf("DeleteApp_%s", uuid.New().String()[:8]),
	}
	require.NoError(t, appRepo.Create(context.Background(), app))

	server := &model.Server{Name: "delete-srv", IP: "10.0.0.2"}
	require.NoError(t, db.Create(server).Error)

	env := &model.Environment{AppID: app.ID, Name: "staging"}
	require.NoError(t, envRepo.Create(context.Background(), env))

	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          "delete-svc",
	}
	require.NoError(t, svcRepo.Create(context.Background(), svc))

	err := appRepo.Delete(context.Background(), app.ID)
	assert.NoError(t, err)

	_, err = appRepo.GetByID(context.Background(), app.ID)
	assert.Error(t, err)
}

func TestAppRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppRepository(db)

	// Count before inserting
	_, totalBefore, err := repo.List(context.Background(), &model.ListAppsRequest{Page: 1, PageSize: 100})
	require.NoError(t, err)

	// Create 3 apps
	for i := 0; i < 3; i++ {
		app := &model.App{
			AppName: fmt.Sprintf("ListApp_%d_%s", i, uuid.New().String()[:8]),
		}
		require.NoError(t, repo.Create(context.Background(), app))
	}

	apps, total, err := repo.List(context.Background(), &model.ListAppsRequest{
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, totalBefore+3, int(total))
	assert.Len(t, apps, 3)
}

func TestAppRepository_ListWithTagFilter(t *testing.T) {
	t.Skip("ILIKE requires PostgreSQL; tag filter uses ILIKE which is PostgreSQL-specific")
}
