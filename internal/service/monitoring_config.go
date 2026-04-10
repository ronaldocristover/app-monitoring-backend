package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/internal/repository"
	"github.com/ronaldocristover/app-monitoring/pkg/apierror"
	"go.uber.org/zap"
)

var (
	ErrMonitoringConfigNotFound = errors.New("monitoring config not found")
)

type MonitoringConfigService interface {
	GetByService(ctx context.Context, serviceID uuid.UUID) (*model.MonitoringConfig, error)
	Upsert(ctx context.Context, serviceID uuid.UUID, req *model.UpdateMonitoringConfigRequest) (*model.MonitoringConfig, error)
}

type monitoringConfigService struct {
	repo        repository.MonitoringConfigRepository
	serviceRepo repository.ServiceRepository
	logger      *zap.SugaredLogger
}

func NewMonitoringConfigService(
	repo repository.MonitoringConfigRepository,
	serviceRepo repository.ServiceRepository,
	logger *zap.SugaredLogger,
) MonitoringConfigService {
	return &monitoringConfigService{
		repo:        repo,
		serviceRepo: serviceRepo,
		logger:      logger,
	}
}

func (s *monitoringConfigService) GetByService(ctx context.Context, serviceID uuid.UUID) (*model.MonitoringConfig, error) {
	config, err := s.repo.GetByService(ctx, serviceID)
	if err != nil {
		return nil, ErrMonitoringConfigNotFound
	}
	return config, nil
}

func (s *monitoringConfigService) Upsert(ctx context.Context, serviceID uuid.UUID, req *model.UpdateMonitoringConfigRequest) (*model.MonitoringConfig, error) {
	if _, err := s.serviceRepo.GetByID(ctx, serviceID); err != nil {
		return nil, ErrServiceNotFound
	}

	existing, _ := s.repo.GetByService(ctx, serviceID)

	config := &model.MonitoringConfig{ServiceID: serviceID}

	if existing != nil {
		config.ID = existing.ID
		config.Enabled = existing.Enabled
		config.PingIntervalSeconds = existing.PingIntervalSeconds
		config.TimeoutSeconds = existing.TimeoutSeconds
		config.Retries = existing.Retries
	}

	if req.Enabled != nil {
		config.Enabled = *req.Enabled
	}
	if req.PingIntervalSeconds != nil {
		config.PingIntervalSeconds = *req.PingIntervalSeconds
	}
	if req.TimeoutSeconds != nil {
		config.TimeoutSeconds = *req.TimeoutSeconds
	}
	if req.Retries != nil {
		config.Retries = *req.Retries
	}

	if err := s.repo.Upsert(ctx, config); err != nil {
		s.logger.Errorw("failed to upsert monitoring config", "error", err)
		return nil, apierror.Internal("Failed to save monitoring config")
	}

	s.logger.Infow("monitoring config upserted", "service_id", serviceID)
	return config, nil
}
