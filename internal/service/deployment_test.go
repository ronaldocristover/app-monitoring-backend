package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/pkg/apierror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeploymentCreate_Success(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	existingService := &model.Service{ID: serviceID}
	req := &model.CreateDeploymentRequest{
		ServiceID:     serviceID,
		Method:        "docker",
		ContainerName: "my-app",
		Port:          8080,
	}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(d *model.Deployment) bool {
		return d.ServiceID == serviceID && d.Method == "docker" && d.Port == 8080
	})).Return(nil)

	result, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, serviceID, result.ServiceID)
	assert.Equal(t, "docker", result.Method)
	assert.Equal(t, "my-app", result.ContainerName)
	assert.Equal(t, 8080, result.Port)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestDeploymentCreate_ServiceNotFound(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	req := &model.CreateDeploymentRequest{
		ServiceID: serviceID,
		Method:    "docker",
	}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(nil, errors.New("not found"))

	result, err := svc.Create(context.Background(), req)

	assert.Nil(t, result)
	assert.Equal(t, ErrServiceNotFound, err)
	svcRepo.AssertExpectations(t)
}

func TestDeploymentCreate_RepoError(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	existingService := &model.Service{ID: serviceID}
	req := &model.CreateDeploymentRequest{
		ServiceID: serviceID,
		Method:    "docker",
	}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	result, err := svc.Create(context.Background(), req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestDeploymentGetByID_Success(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	expected := &model.Deployment{ID: id, Method: "docker"}

	repo.On("GetByID", mock.Anything, id).Return(expected, nil)

	result, err := svc.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestDeploymentGetByID_NotFound(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	result, err := svc.GetByID(context.Background(), id)

	assert.Nil(t, result)
	assert.Equal(t, ErrDeploymentNotFound, err)
	repo.AssertExpectations(t)
}

func TestDeploymentUpdate_Success(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Deployment{ID: id, Method: "docker", ContainerName: "old", Port: 8080}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(d *model.Deployment) bool {
		return d.Method == "k8s" && d.Port == 9090
	})).Return(nil)

	req := &model.UpdateDeploymentRequest{
		Method: "k8s",
		Port:   9090,
	}
	result, err := svc.Update(context.Background(), id, req)

	assert.NoError(t, err)
	assert.Equal(t, "k8s", result.Method)
	assert.Equal(t, 9090, result.Port)
	repo.AssertExpectations(t)
}

func TestDeploymentUpdate_NotFound(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	req := &model.UpdateDeploymentRequest{Method: "k8s"}
	result, err := svc.Update(context.Background(), id, req)

	assert.Nil(t, result)
	assert.Equal(t, ErrDeploymentNotFound, err)
	repo.AssertExpectations(t)
}

func TestDeploymentUpdate_RepoError(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Deployment{ID: id, Method: "docker"}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))

	req := &model.UpdateDeploymentRequest{Method: "k8s"}
	result, err := svc.Update(context.Background(), id, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	repo.AssertExpectations(t)
}

func TestDeploymentDelete_Success(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Deployment{ID: id}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Delete", mock.Anything, id).Return(nil)

	err := svc.Delete(context.Background(), id)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeploymentDelete_NotFound(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	err := svc.Delete(context.Background(), id)

	assert.Equal(t, ErrDeploymentNotFound, err)
	repo.AssertExpectations(t)
}

func TestDeploymentDelete_RepoError(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Deployment{ID: id}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Delete", mock.Anything, id).Return(errors.New("db error"))

	err := svc.Delete(context.Background(), id)

	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	repo.AssertExpectations(t)
}

func TestDeploymentListByService_Success(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	existingService := &model.Service{ID: serviceID}
	deployments := []*model.Deployment{
		{ID: uuid.New(), ServiceID: serviceID, Method: "docker"},
		{ID: uuid.New(), ServiceID: serviceID, Method: "k8s"},
	}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("ListByService", mock.Anything, serviceID, mock.Anything).Return(deployments, int64(2), nil)

	result, total, err := svc.ListByService(context.Background(), serviceID, &model.ListDeploymentsRequest{})

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestDeploymentListByService_ServiceNotFound(t *testing.T) {
	repo := new(MockDeploymentRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewDeploymentService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	svcRepo.On("GetByID", mock.Anything, serviceID).Return(nil, errors.New("not found"))

	result, total, err := svc.ListByService(context.Background(), serviceID, &model.ListDeploymentsRequest{})

	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, ErrServiceNotFound, err)
	svcRepo.AssertExpectations(t)
}
