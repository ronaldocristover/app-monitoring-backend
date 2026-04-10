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

func TestMonitoringConfigGetByService_Success(t *testing.T) {
	repo := new(MockMonitoringConfigRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewMonitoringConfigService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	expected := &model.MonitoringConfig{
		ID:        uuid.New(),
		ServiceID: serviceID,
		Enabled:   true,
	}

	repo.On("GetByService", mock.Anything, serviceID).Return(expected, nil)

	result, err := svc.GetByService(context.Background(), serviceID)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestMonitoringConfigGetByService_NotFound(t *testing.T) {
	repo := new(MockMonitoringConfigRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewMonitoringConfigService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	repo.On("GetByService", mock.Anything, serviceID).Return(nil, errors.New("not found"))

	result, err := svc.GetByService(context.Background(), serviceID)

	assert.Nil(t, result)
	assert.Equal(t, ErrMonitoringConfigNotFound, err)
	repo.AssertExpectations(t)
}

func TestMonitoringConfigUpsert_CreateNew(t *testing.T) {
	repo := new(MockMonitoringConfigRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewMonitoringConfigService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	existingService := &model.Service{ID: serviceID}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("GetByService", mock.Anything, serviceID).Return(nil, errors.New("not found"))
	repo.On("Upsert", mock.Anything, mock.MatchedBy(func(c *model.MonitoringConfig) bool {
		return c.ServiceID == serviceID && c.Enabled
	})).Return(nil)

	enabled := true
	interval := 30
	req := &model.UpdateMonitoringConfigRequest{
		Enabled:             &enabled,
		PingIntervalSeconds: &interval,
	}

	result, err := svc.Upsert(context.Background(), serviceID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Enabled)
	assert.Equal(t, 30, result.PingIntervalSeconds)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestMonitoringConfigUpsert_UpdateExisting(t *testing.T) {
	repo := new(MockMonitoringConfigRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewMonitoringConfigService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	configID := uuid.New()
	existingService := &model.Service{ID: serviceID}
	existingConfig := &model.MonitoringConfig{
		ID:                  configID,
		ServiceID:           serviceID,
		Enabled:             true,
		PingIntervalSeconds: 60,
		TimeoutSeconds:      10,
		Retries:             3,
	}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("GetByService", mock.Anything, serviceID).Return(existingConfig, nil)
	repo.On("Upsert", mock.Anything, mock.MatchedBy(func(c *model.MonitoringConfig) bool {
		return c.ID == configID && c.PingIntervalSeconds == 15
	})).Return(nil)

	interval := 15
	req := &model.UpdateMonitoringConfigRequest{
		PingIntervalSeconds: &interval,
	}

	result, err := svc.Upsert(context.Background(), serviceID, req)

	assert.NoError(t, err)
	assert.Equal(t, configID, result.ID)
	assert.Equal(t, 15, result.PingIntervalSeconds)
	// Preserves existing values when not updated
	assert.True(t, result.Enabled)
	assert.Equal(t, 10, result.TimeoutSeconds)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestMonitoringConfigUpsert_ServiceNotFound(t *testing.T) {
	repo := new(MockMonitoringConfigRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewMonitoringConfigService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	svcRepo.On("GetByID", mock.Anything, serviceID).Return(nil, errors.New("not found"))

	req := &model.UpdateMonitoringConfigRequest{}
	result, err := svc.Upsert(context.Background(), serviceID, req)

	assert.Nil(t, result)
	assert.Equal(t, ErrServiceNotFound, err)
	svcRepo.AssertExpectations(t)
}

func TestMonitoringConfigUpsert_RepoError(t *testing.T) {
	repo := new(MockMonitoringConfigRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewMonitoringConfigService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	existingService := &model.Service{ID: serviceID}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("GetByService", mock.Anything, serviceID).Return(nil, errors.New("not found"))
	repo.On("Upsert", mock.Anything, mock.Anything).Return(errors.New("db error"))

	req := &model.UpdateMonitoringConfigRequest{}
	result, err := svc.Upsert(context.Background(), serviceID, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}
