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

func createTestApp(t *testing.T, db interface{ Create(value interface{}) interface{ Error() error } }) *model.App {
	t.Helper()
	app := &model.App{AppName: "Test App " + uuid.New().String()}
	require.NoError(t, db.Create(app).Error())
	return app
}

func TestEnvironmentRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)
	ctx := context.Background()

	app := &model.App{AppName: "App for Env Create"}
	require.NoError(t, appRepo.Create(ctx, app))

	env := &model.Environment{
		AppID: app.ID,
		Name:  "production",
	}

	err := envRepo.Create(ctx, env)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, env.ID)
}

func TestEnvironmentRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)
	ctx := context.Background()

	app := &model.App{AppName: "App for Env GetByID"}
	require.NoError(t, appRepo.Create(ctx, app))

	env := &model.Environment{AppID: app.ID, Name: "staging"}
	require.NoError(t, envRepo.Create(ctx, env))

	found, err := envRepo.GetByID(ctx, env.ID)
	assert.NoError(t, err)
	assert.Equal(t, env.ID, found.ID)
	assert.Equal(t, "staging", found.Name)
	assert.Equal(t, app.ID, found.AppID)

	// Non-existent ID returns error
	_, err = envRepo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
}

func TestEnvironmentRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)
	ctx := context.Background()

	app := &model.App{AppName: "App for Env Update"}
	require.NoError(t, appRepo.Create(ctx, app))

	env := &model.Environment{AppID: app.ID, Name: "staging"}
	require.NoError(t, envRepo.Create(ctx, env))

	env.Name = "production"
	err := envRepo.Update(ctx, env)
	assert.NoError(t, err)

	found, err := envRepo.GetByID(ctx, env.ID)
	assert.NoError(t, err)
	assert.Equal(t, "production", found.Name)
}

func TestEnvironmentRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)
	ctx := context.Background()

	app := &model.App{AppName: "App for Env Delete"}
	require.NoError(t, appRepo.Create(ctx, app))

	env := &model.Environment{AppID: app.ID, Name: "staging"}
	require.NoError(t, envRepo.Create(ctx, env))

	err := envRepo.Delete(ctx, env.ID)
	assert.NoError(t, err)

	_, err = envRepo.GetByID(ctx, env.ID)
	assert.Error(t, err)
}

func TestEnvironmentRepository_List(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)
	ctx := context.Background()

	app := &model.App{AppName: "App for Env List"}
	require.NoError(t, appRepo.Create(ctx, app))

	_, beforeCount, err := envRepo.List(ctx, &model.ListEnvironmentsRequest{Page: 1, PageSize: 100})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		require.NoError(t, envRepo.Create(ctx, &model.Environment{
			AppID: app.ID,
			Name:  fmt.Sprintf("env-%d", i),
		}))
	}

	envs, total, err := envRepo.List(ctx, &model.ListEnvironmentsRequest{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, beforeCount+3, total)
	assert.Equal(t, beforeCount+3, int64(len(envs)))
}

func TestEnvironmentRepository_ListByApp(t *testing.T) {
	db := setupTestDB(t)
	appRepo := NewAppRepository(db)
	envRepo := NewEnvironmentRepository(db)
	ctx := context.Background()

	app1 := &model.App{AppName: "App One"}
	app2 := &model.App{AppName: "App Two"}
	require.NoError(t, appRepo.Create(ctx, app1))
	require.NoError(t, appRepo.Create(ctx, app2))

	// 2 envs for app1
	require.NoError(t, envRepo.Create(ctx, &model.Environment{AppID: app1.ID, Name: "prod"}))
	require.NoError(t, envRepo.Create(ctx, &model.Environment{AppID: app1.ID, Name: "staging"}))

	// 1 env for app2
	require.NoError(t, envRepo.Create(ctx, &model.Environment{AppID: app2.ID, Name: "dev"}))

	filter := &model.ListEnvironmentsRequest{
		Page:     1,
		PageSize: 10,
		AppID:    app1.ID.String(),
	}
	envs, total, err := envRepo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, envs, 2)
}
