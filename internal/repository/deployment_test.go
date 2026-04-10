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

func createTestServiceForDeployment(t *testing.T, db *gorm.DB) *model.Service {
	t.Helper()
	app := &model.App{AppName: "DeployApp_" + t.Name()}
	require.NoError(t, db.Create(app).Error)
	server := &model.Server{Name: "DeployServer_" + t.Name(), IP: "10.0.0.3"}
	require.NoError(t, db.Create(server).Error)
	env := &model.Environment{AppID: app.ID, Name: "production_" + t.Name()}
	require.NoError(t, db.Create(env).Error)
	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          "DeployService_" + t.Name(),
	}
	require.NoError(t, db.Create(svc).Error)
	return svc
}

func TestDeploymentRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)
	ctx := context.Background()

	svc := createTestServiceForDeployment(t, db)

	deployment := &model.Deployment{
		ServiceID:     svc.ID,
		Method:        "docker",
		ContainerName: "my-container",
		Port:          8080,
		Config:        `{"image": "nginx:latest"}`,
	}

	err := repo.Create(ctx, deployment)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, deployment.ID)
}

func TestDeploymentRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)
	ctx := context.Background()

	svc := createTestServiceForDeployment(t, db)

	deployment := &model.Deployment{
		ServiceID:     svc.ID,
		Method:        "docker",
		ContainerName: "test-container",
		Port:          8080,
		Config:        `{"image": "nginx:latest"}`,
	}
	require.NoError(t, repo.Create(ctx, deployment))

	found, err := repo.GetByID(ctx, deployment.ID)
	assert.NoError(t, err)
	assert.Equal(t, deployment.ID, found.ID)
	assert.Equal(t, svc.ID, found.ServiceID)
	assert.Equal(t, "docker", found.Method)
	assert.Equal(t, "test-container", found.ContainerName)
	assert.Equal(t, 8080, found.Port)

	// Test non-existent ID
	_, err = repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
}

func TestDeploymentRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)
	ctx := context.Background()

	svc := createTestServiceForDeployment(t, db)

	deployment := &model.Deployment{
		ServiceID:     svc.ID,
		Method:        "docker",
		ContainerName: "original-container",
		Port:          8080,
	}
	require.NoError(t, repo.Create(ctx, deployment))

	deployment.Method = "kubernetes"
	deployment.Port = 9090
	err := repo.Update(ctx, deployment)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, deployment.ID)
	assert.NoError(t, err)
	assert.Equal(t, "kubernetes", found.Method)
	assert.Equal(t, 9090, found.Port)
}

func TestDeploymentRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)
	ctx := context.Background()

	svc := createTestServiceForDeployment(t, db)

	deployment := &model.Deployment{
		ServiceID:     svc.ID,
		Method:        "docker",
		ContainerName: "delete-container",
		Port:          8080,
	}
	require.NoError(t, repo.Create(ctx, deployment))

	err := repo.Delete(ctx, deployment.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, deployment.ID)
	assert.Error(t, err)
}

func TestDeploymentRepository_ListByService(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)
	ctx := context.Background()

	svc := createTestServiceForDeployment(t, db)

	// Count before inserting
	var countBefore int64
	db.Model(&model.Deployment{}).Where("service_id = ?", svc.ID).Count(&countBefore)

	// Create 3 deployments for the same service
	for i := 0; i < 3; i++ {
		deployment := &model.Deployment{
			ServiceID:     svc.ID,
			Method:        "docker",
			ContainerName: fmt.Sprintf("container_%d", i),
			Port:          8080 + i,
		}
		require.NoError(t, repo.Create(ctx, deployment))
	}

	deployments, total, err := repo.ListByService(ctx, svc.ID, &model.ListDeploymentsRequest{
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, countBefore+3, total)
	assert.Len(t, deployments, 3)
}
