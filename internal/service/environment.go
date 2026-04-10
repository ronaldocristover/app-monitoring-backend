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
	ErrEnvironmentNotFound = errors.New("environment not found")
	ErrInvalidAppID        = errors.New("invalid app ID")
)

type EnvironmentService interface {
	Create(ctx context.Context, req *model.CreateEnvironmentRequest) (*model.Environment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Environment, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateEnvironmentRequest) (*model.Environment, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListEnvironmentsRequest) ([]*model.Environment, int64, error)
}

type environmentService struct {
	repo   repository.EnvironmentRepository
	logger *zap.SugaredLogger
}

func NewEnvironmentService(repo repository.EnvironmentRepository, logger *zap.SugaredLogger) EnvironmentService {
	return &environmentService{repo: repo, logger: logger}
}

func (s *environmentService) Create(ctx context.Context, req *model.CreateEnvironmentRequest) (*model.Environment, error) {
	env := &model.Environment{
		AppID: req.AppID,
		Name:  req.Name,
	}
	if err := s.repo.Create(ctx, env); err != nil {
		s.logger.Errorw("failed to create environment", "error", err)
		return nil, apierror.Internal("Failed to create environment")
	}
	return env, nil
}

func (s *environmentService) GetByID(ctx context.Context, id uuid.UUID) (*model.Environment, error) {
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}
	return env, nil
}

func (s *environmentService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateEnvironmentRequest) (*model.Environment, error) {
	env, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}
	if req.Name != "" {
		env.Name = req.Name
	}
	if err := s.repo.Update(ctx, env); err != nil {
		s.logger.Errorw("failed to update environment", "error", err)
		return nil, apierror.Internal("Failed to update environment")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *environmentService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrEnvironmentNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete environment", "error", err)
		return apierror.Internal("Failed to delete environment")
	}
	return nil
}

func (s *environmentService) List(ctx context.Context, req *model.ListEnvironmentsRequest) ([]*model.Environment, int64, error) {
	return s.repo.List(ctx, req)
}
