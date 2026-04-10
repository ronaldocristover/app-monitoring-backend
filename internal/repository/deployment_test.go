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

func createTestDeployment(t *testing.T, db *gorm.DB, repo DeploymentRepository, suffix string) (*model.Service, *model.Deployment) {
	t.Helper()
	app := &model.App{AppName: fmt.Sprintf("DeployApp_%s", suffix)}
	require.NoError(t, db.Create(app).Error)

	env := &model.Environment{AppID: app.ID, Name: "prod"}
	require.NoError(t, db.Create(env).Error)

	server := &model.Server{Name: fmt.Sprintf("srv_%s", suffix), IP: "10.0.0.1"}
	require.NoError(t, db.Create(server).Error)

	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          fmt.Sprintf("deploy-svc-%s", suffix),
	}
	require.NoError(t, db.Create(svc).Error)

	deployment := &model.Deployment{
		ServiceID:     svc.ID,
		Method:        "docker",
		ContainerName: fmt.Sprintf("container_%s", suffix),
		Port:          8080,
		Config:        `{"image": "nginx:latest"}`,
	}
	require.NoError(t, repo.Create(context.Background(), deployment))
	return svc, deployment
}

func TestDeploymentRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)

	_, deployment := createTestDeployment(t, db, repo, "create")
	assert.NotEqual(t, uuid.Nil, deployment.ID)
}

func TestDeploymentRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)

	_, deployment := createTestDeployment(t, db, repo, "get")

	found, err := repo.GetByID(context.Background(), deployment.ID)
	assert.NoError(t, err)
	assert.Equal(t, deployment.ID, found.ID)
	assert.Equal(t, deployment.ServiceID, found.ServiceID)
	assert.Equal(t, deployment.Method, found.Method)
	assert.Equal(t, deployment.ContainerName, found.ContainerName)
	assert.Equal(t, deployment.Port, found.Port)
}

func TestDeploymentRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestDeploymentRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)

	_, deployment := createTestDeployment(t, db, repo, "update")
	deployment.Method = "kubernetes"
	deployment.Port = 9090

	err := repo.Update(context.Background(), deployment)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), deployment.ID)
	assert.NoError(t, err)
	assert.Equal(t, "kubernetes", found.Method)
	assert.Equal(t, 9090, found.Port)
}

func TestDeploymentRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)

	_, deployment := createTestDeployment(t, db, repo, "delete")

	err := repo.Delete(context.Background(), deployment.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), deployment.ID)
	assert.Error(t, err)
}

func TestDeploymentRepository_ListByService(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)

	svc, _ := createTestDeployment(t, db, repo, "list1")
	_, _ = createTestDeployment(t, db, repo, "list2")

	// Create another deployment for the same service
	deployment3 := &model.Deployment{
		ServiceID:     svc.ID,
		Method:        "kubernetes",
		ContainerName: "container_extra",
		Port:          3000,
	}
	require.NoError(t, repo.Create(context.Background(), deployment3))

	deployments, total, err := repo.ListByService(context.Background(), svc.ID, &model.ListDeploymentsRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, deployments, 2)
}

func TestDeploymentRepository_ListByService_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)

	svc, _ := createTestDeployment(t, db, repo, "page")

	for i := 0; i < 4; i++ {
		d := &model.Deployment{
			ServiceID:     svc.ID,
			Method:        "docker",
			ContainerName: fmt.Sprintf("cnt_%d", i),
			Port:          8080 + i,
		}
		require.NoError(t, repo.Create(context.Background(), d))
	}

	deployments, total, err := repo.ListByService(context.Background(), svc.ID, &model.ListDeploymentsRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total) // 1 from createTestDeployment + 4
	assert.Len(t, deployments, 2)
}

func TestDeploymentRepository_ListByService_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)

	// Create a service with no deployments
	app := &model.App{AppName: "EmptyDeployApp"}
	require.NoError(t, db.Create(app).Error)
	env := &model.Environment{AppID: app.ID, Name: "prod"}
	require.NoError(t, db.Create(env).Error)
	server := &model.Server{Name: "srv", IP: "10.0.0.1"}
	require.NoError(t, db.Create(server).Error)
	svc := &model.Service{EnvironmentID: env.ID, ServerID: server.ID, Name: "empty-svc"}
	require.NoError(t, db.Create(svc).Error)

	deployments, total, err := repo.ListByService(context.Background(), svc.ID, &model.ListDeploymentsRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, deployments, 0)
}

func TestNewDeploymentRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeploymentRepository(db)
	assert.NotNil(t, repo)
}
