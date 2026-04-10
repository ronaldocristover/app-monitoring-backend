package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/pkg/apierror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMonitoringLogListByService_Success(t *testing.T) {
	repo := new(MockMonitoringLogRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewMonitoringLogService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	existingService := &model.Service{ID: serviceID}
	logs := []*model.MonitoringLog{
		{ID: uuid.New(), ServiceID: serviceID, Status: "up"},
		{ID: uuid.New(), ServiceID: serviceID, Status: "down"},
	}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("ListByService", mock.Anything, serviceID, mock.Anything).Return(logs, int64(2), nil)

	result, total, err := svc.ListByService(context.Background(), serviceID, &model.ListMonitoringLogsRequest{})

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestMonitoringLogListByService_ServiceNotFound(t *testing.T) {
	repo := new(MockMonitoringLogRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewMonitoringLogService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	svcRepo.On("GetByID", mock.Anything, serviceID).Return(nil, errors.New("not found"))

	result, total, err := svc.ListByService(context.Background(), serviceID, &model.ListMonitoringLogsRequest{})

	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, ErrServiceNotFound, err)
	svcRepo.AssertExpectations(t)
}

func TestMonitoringLogListByService_RepoError(t *testing.T) {
	repo := new(MockMonitoringLogRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewMonitoringLogService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	existingService := &model.Service{ID: serviceID}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("ListByService", mock.Anything, serviceID, mock.Anything).Return([]*model.MonitoringLog{}, int64(0), errors.New("db error"))

	result, total, err := svc.ListByService(context.Background(), serviceID, &model.ListMonitoringLogsRequest{})

	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestMonitoringLogGetLatest_Success(t *testing.T) {
	repo := new(MockMonitoringLogRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewMonitoringLogService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	existingService := &model.Service{ID: serviceID}
	expected := &model.MonitoringLog{
		ID:         uuid.New(),
		ServiceID:  serviceID,
		Status:     "up",
		CheckedAt:  time.Now(),
	}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("GetLatest", mock.Anything, serviceID).Return(expected, nil)

	result, err := svc.GetLatest(context.Background(), serviceID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestMonitoringLogGetLatest_ServiceNotFound(t *testing.T) {
	repo := new(MockMonitoringLogRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewMonitoringLogService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	svcRepo.On("GetByID", mock.Anything, serviceID).Return(nil, errors.New("not found"))

	result, err := svc.GetLatest(context.Background(), serviceID)

	assert.Nil(t, result)
	assert.Equal(t, ErrServiceNotFound, err)
	svcRepo.AssertExpectations(t)
}

func TestMonitoringLogGetLatest_NotFound(t *testing.T) {
	repo := new(MockMonitoringLogRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewMonitoringLogService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	existingService := &model.Service{ID: serviceID}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("GetLatest", mock.Anything, serviceID).Return(nil, errors.New("not found"))

	result, err := svc.GetLatest(context.Background(), serviceID)

	assert.Nil(t, result)
	assert.Equal(t, ErrMonitoringLogNotFound, err)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}
