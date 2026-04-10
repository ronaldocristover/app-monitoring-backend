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
	ErrDeploymentNotFound = errors.New("deployment not found")
)

type DeploymentService interface {
	Create(ctx context.Context, req *model.CreateDeploymentRequest) (*model.Deployment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Deployment, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateDeploymentRequest) (*model.Deployment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByService(ctx context.Context, serviceID uuid.UUID, req *model.ListDeploymentsRequest) ([]*model.Deployment, int64, error)
}

type deploymentService struct {
	repo        repository.DeploymentRepository
	serviceRepo repository.ServiceRepository
	logger      *zap.SugaredLogger
}

func NewDeploymentService(
	repo repository.DeploymentRepository,
	serviceRepo repository.ServiceRepository,
	logger *zap.SugaredLogger,
) DeploymentService {
	return &deploymentService{
		repo:        repo,
		serviceRepo: serviceRepo,
		logger:      logger,
	}
}

func (s *deploymentService) Create(ctx context.Context, req *model.CreateDeploymentRequest) (*model.Deployment, error) {
	if _, err := s.serviceRepo.GetByID(ctx, req.ServiceID); err != nil {
		return nil, ErrServiceNotFound
	}

	deployment := &model.Deployment{
		ServiceID:     req.ServiceID,
		Method:        req.Method,
		ContainerName: req.ContainerName,
		Port:          req.Port,
		Config:        req.Config,
	}

	if err := s.repo.Create(ctx, deployment); err != nil {
		s.logger.Errorw("failed to create deployment", "error", err)
		return nil, apierror.Internal("Failed to create deployment")
	}

	s.logger.Infow("deployment created", "id", deployment.ID)
	return deployment, nil
}

func (s *deploymentService) GetByID(ctx context.Context, id uuid.UUID) (*model.Deployment, error) {
	deployment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrDeploymentNotFound
	}
	return deployment, nil
}

func (s *deploymentService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateDeploymentRequest) (*model.Deployment, error) {
	deployment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrDeploymentNotFound
	}

	if req.Method != "" {
		deployment.Method = req.Method
	}
	if req.ContainerName != "" {
		deployment.ContainerName = req.ContainerName
	}
	if req.Port != 0 {
		deployment.Port = req.Port
	}
	if req.Config != "" {
		deployment.Config = req.Config
	}

	if err := s.repo.Update(ctx, deployment); err != nil {
		s.logger.Errorw("failed to update deployment", "error", err)
		return nil, apierror.Internal("Failed to update deployment")
	}

	return deployment, nil
}

func (s *deploymentService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return ErrDeploymentNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete deployment", "error", err)
		return apierror.Internal("Failed to delete deployment")
	}
	return nil
}

func (s *deploymentService) ListByService(ctx context.Context, serviceID uuid.UUID, req *model.ListDeploymentsRequest) ([]*model.Deployment, int64, error) {
	if _, err := s.serviceRepo.GetByID(ctx, serviceID); err != nil {
		return nil, 0, ErrServiceNotFound
	}

	return s.repo.ListByService(ctx, serviceID, req)
}
