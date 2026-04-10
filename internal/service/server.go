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
	ErrServerNotFound = errors.New("server not found")
)

type ServerService interface {
	Create(ctx context.Context, req *model.CreateServerRequest) (*model.Server, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Server, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateServerRequest) (*model.Server, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListServersRequest) ([]*model.Server, int64, error)
}

type serverService struct {
	repo   repository.ServerRepository
	logger *zap.SugaredLogger
}

func NewServerService(repo repository.ServerRepository, logger *zap.SugaredLogger) ServerService {
	return &serverService{repo: repo, logger: logger}
}

func (s *serverService) Create(ctx context.Context, req *model.CreateServerRequest) (*model.Server, error) {
	server := &model.Server{
		Name:     req.Name,
		IP:       req.IP,
		Provider: req.Provider,
	}
	if err := s.repo.Create(ctx, server); err != nil {
		s.logger.Errorw("failed to create server", "error", err)
		return nil, apierror.Internal("Failed to create server")
	}
	return server, nil
}

func (s *serverService) GetByID(ctx context.Context, id uuid.UUID) (*model.Server, error) {
	server, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrServerNotFound
	}
	return server, nil
}

func (s *serverService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateServerRequest) (*model.Server, error) {
	server, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrServerNotFound
	}
	if req.Name != "" {
		server.Name = req.Name
	}
	if req.IP != "" {
		server.IP = req.IP
	}
	if req.Provider != "" {
		server.Provider = req.Provider
	}
	if err := s.repo.Update(ctx, server); err != nil {
		s.logger.Errorw("failed to update server", "error", err)
		return nil, apierror.Internal("Failed to update server")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *serverService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrServerNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete server", "error", err)
		return apierror.Internal("Failed to delete server")
	}
	return nil
}

func (s *serverService) List(ctx context.Context, req *model.ListServersRequest) ([]*model.Server, int64, error) {
	return s.repo.List(ctx, req)
}
