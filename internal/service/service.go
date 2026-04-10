package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/internal/repository"
	"github.com/ronaldocristover/app-monitoring/pkg/apierror"
	"go.uber.org/zap"
)

var (
	ErrServiceNotFound = errors.New("service not found")
)

const defaultPingTimeout = 10 * time.Second

type ServiceService interface {
	Create(ctx context.Context, req *model.CreateServiceRequest) (*model.Service, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Service, error)
	GetByIDFull(ctx context.Context, id uuid.UUID) (*model.Service, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateServiceRequest) (*model.Service, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListServicesRequest) ([]*model.Service, int64, error)
	ManualPing(ctx context.Context, id uuid.UUID) (*model.MonitoringLog, error)
}

type serviceService struct {
	repo          repository.ServiceRepository
	monitoringLog repository.MonitoringLogRepository
	logger        *zap.SugaredLogger
}

func NewServiceService(
	repo repository.ServiceRepository,
	monitoringLog repository.MonitoringLogRepository,
	logger *zap.SugaredLogger,
) ServiceService {
	return &serviceService{
		repo:          repo,
		monitoringLog: monitoringLog,
		logger:        logger,
	}
}

func (s *serviceService) Create(ctx context.Context, req *model.CreateServiceRequest) (*model.Service, error) {
	svc := &model.Service{
		EnvironmentID:  req.EnvironmentID,
		ServerID:       req.ServerID,
		Name:           req.Name,
		Type:           req.Type,
		URL:            req.URL,
		Repository:     req.Repository,
		StackLanguage:  req.StackLanguage,
		StackFramework: req.StackFramework,
		DBType:         req.DBType,
		DBHost:         req.DBHost,
	}

	if err := s.repo.Create(ctx, svc); err != nil {
		s.logger.Errorw("failed to create service", "error", err)
		return nil, apierror.Internal("Failed to create service")
	}

	s.logger.Infow("service created", "id", svc.ID)
	return svc, nil
}

func (s *serviceService) GetByID(ctx context.Context, id uuid.UUID) (*model.Service, error) {
	svc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrServiceNotFound
	}
	return svc, nil
}

func (s *serviceService) GetByIDFull(ctx context.Context, id uuid.UUID) (*model.Service, error) {
	svc, err := s.repo.GetByIDFull(ctx, id)
	if err != nil {
		return nil, ErrServiceNotFound
	}
	return svc, nil
}

func (s *serviceService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateServiceRequest) (*model.Service, error) {
	svc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrServiceNotFound
	}

	if req.EnvironmentID != nil {
		svc.EnvironmentID = *req.EnvironmentID
	}
	if req.ServerID != nil {
		svc.ServerID = *req.ServerID
	}
	if req.Name != "" {
		svc.Name = req.Name
	}
	if req.Type != "" {
		svc.Type = req.Type
	}
	if req.URL != "" {
		svc.URL = req.URL
	}
	if req.Repository != "" {
		svc.Repository = req.Repository
	}
	if req.StackLanguage != "" {
		svc.StackLanguage = req.StackLanguage
	}
	if req.StackFramework != "" {
		svc.StackFramework = req.StackFramework
	}
	if req.DBType != "" {
		svc.DBType = req.DBType
	}
	if req.DBHost != "" {
		svc.DBHost = req.DBHost
	}

	if err := s.repo.Update(ctx, svc); err != nil {
		s.logger.Errorw("failed to update service", "error", err)
		return nil, apierror.Internal("Failed to update service")
	}

	return svc, nil
}

func (s *serviceService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return ErrServiceNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete service", "error", err)
		return apierror.Internal("Failed to delete service")
	}
	return nil
}

func (s *serviceService) List(ctx context.Context, req *model.ListServicesRequest) ([]*model.Service, int64, error) {
	return s.repo.List(ctx, req)
}

func (s *serviceService) ManualPing(ctx context.Context, id uuid.UUID) (*model.MonitoringLog, error) {
	svc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrServiceNotFound
	}

	if svc.URL == "" {
		return nil, apierror.BadRequest("Service has no URL configured")
	}

	log := &model.MonitoringLog{
		ServiceID: id,
		CheckedAt: time.Now(),
	}

	client := &http.Client{Timeout: defaultPingTimeout}
	start := time.Now()
	resp, err := client.Get(svc.URL)
	elapsed := time.Since(start)
	log.ResponseTimeMs = int(elapsed.Milliseconds())

	if err != nil {
		log.Status = "down"
		log.ErrorMessage = err.Error()
	} else {
		defer resp.Body.Close()
		log.StatusCode = resp.StatusCode
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			log.Status = "up"
		} else {
			log.Status = "down"
			log.ErrorMessage = http.StatusText(resp.StatusCode)
		}
	}

	if err := s.monitoringLog.Create(ctx, log); err != nil {
		s.logger.Errorw("failed to save monitoring log", "error", err)
		return nil, apierror.Internal("Failed to save monitoring log")
	}

	s.logger.Infow("manual ping completed", "service_id", id, "status", log.Status, "response_ms", log.ResponseTimeMs)
	return log, nil
}
