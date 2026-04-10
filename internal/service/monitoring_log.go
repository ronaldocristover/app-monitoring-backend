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
	ErrMonitoringLogNotFound = errors.New("monitoring log not found")
)

type MonitoringLogService interface {
	ListByService(ctx context.Context, serviceID uuid.UUID, req *model.ListMonitoringLogsRequest) ([]*model.MonitoringLog, int64, error)
	GetLatest(ctx context.Context, serviceID uuid.UUID) (*model.MonitoringLog, error)
}

type monitoringLogService struct {
	repo        repository.MonitoringLogRepository
	serviceRepo repository.ServiceRepository
	logger      *zap.SugaredLogger
}

func NewMonitoringLogService(
	repo repository.MonitoringLogRepository,
	serviceRepo repository.ServiceRepository,
	logger *zap.SugaredLogger,
) MonitoringLogService {
	return &monitoringLogService{
		repo:        repo,
		serviceRepo: serviceRepo,
		logger:      logger,
	}
}

func (s *monitoringLogService) ListByService(ctx context.Context, serviceID uuid.UUID, req *model.ListMonitoringLogsRequest) ([]*model.MonitoringLog, int64, error) {
	if _, err := s.serviceRepo.GetByID(ctx, serviceID); err != nil {
		return nil, 0, ErrServiceNotFound
	}

	logs, total, err := s.repo.ListByService(ctx, serviceID, req)
	if err != nil {
		s.logger.Errorw("failed to list monitoring logs", "error", err)
		return nil, 0, apierror.Internal("Failed to fetch monitoring logs")
	}

	return logs, total, nil
}

func (s *monitoringLogService) GetLatest(ctx context.Context, serviceID uuid.UUID) (*model.MonitoringLog, error) {
	if _, err := s.serviceRepo.GetByID(ctx, serviceID); err != nil {
		return nil, ErrServiceNotFound
	}

	log, err := s.repo.GetLatest(ctx, serviceID)
	if err != nil {
		return nil, ErrMonitoringLogNotFound
	}
	return log, nil
}
