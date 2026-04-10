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
	ErrBackupNotFound = errors.New("backup not found")
)

type BackupService interface {
	Create(ctx context.Context, req *model.CreateBackupRequest) (*model.Backup, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Backup, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateBackupRequest) (*model.Backup, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByService(ctx context.Context, serviceID uuid.UUID, req *model.ListBackupsRequest) ([]*model.Backup, int64, error)
}

type backupService struct {
	repo        repository.BackupRepository
	serviceRepo repository.ServiceRepository
	logger      *zap.SugaredLogger
}

func NewBackupService(
	repo repository.BackupRepository,
	serviceRepo repository.ServiceRepository,
	logger *zap.SugaredLogger,
) BackupService {
	return &backupService{
		repo:        repo,
		serviceRepo: serviceRepo,
		logger:      logger,
	}
}

func (s *backupService) Create(ctx context.Context, req *model.CreateBackupRequest) (*model.Backup, error) {
	if _, err := s.serviceRepo.GetByID(ctx, req.ServiceID); err != nil {
		return nil, ErrServiceNotFound
	}

	backup := &model.Backup{
		ServiceID: req.ServiceID,
		Enabled:   req.Enabled,
		Path:      req.Path,
		Schedule:  req.Schedule,
	}

	if err := s.repo.Create(ctx, backup); err != nil {
		s.logger.Errorw("failed to create backup", "error", err)
		return nil, apierror.Internal("Failed to create backup")
	}

	s.logger.Infow("backup created", "id", backup.ID)
	return backup, nil
}

func (s *backupService) GetByID(ctx context.Context, id uuid.UUID) (*model.Backup, error) {
	backup, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrBackupNotFound
	}
	return backup, nil
}

func (s *backupService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateBackupRequest) (*model.Backup, error) {
	backup, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrBackupNotFound
	}

	if req.Enabled != nil {
		backup.Enabled = *req.Enabled
	}
	if req.Path != "" {
		backup.Path = req.Path
	}
	if req.Schedule != "" {
		backup.Schedule = req.Schedule
	}
	if req.Status != "" {
		backup.Status = req.Status
	}

	if err := s.repo.Update(ctx, backup); err != nil {
		s.logger.Errorw("failed to update backup", "error", err)
		return nil, apierror.Internal("Failed to update backup")
	}

	return backup, nil
}

func (s *backupService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return ErrBackupNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete backup", "error", err)
		return apierror.Internal("Failed to delete backup")
	}
	return nil
}

func (s *backupService) ListByService(ctx context.Context, serviceID uuid.UUID, req *model.ListBackupsRequest) ([]*model.Backup, int64, error) {
	if _, err := s.serviceRepo.GetByID(ctx, serviceID); err != nil {
		return nil, 0, ErrServiceNotFound
	}

	return s.repo.ListByService(ctx, serviceID, req)
}
