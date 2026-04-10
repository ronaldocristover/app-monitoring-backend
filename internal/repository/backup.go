package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/pkg/pagination"
	"gorm.io/gorm"
)

type BackupRepository interface {
	Create(ctx context.Context, backup *model.Backup) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Backup, error)
	Update(ctx context.Context, backup *model.Backup) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByService(ctx context.Context, serviceID uuid.UUID, filter *model.ListBackupsRequest) ([]*model.Backup, int64, error)
}

type backupRepository struct {
	db *gorm.DB
}

func NewBackupRepository(db *gorm.DB) BackupRepository {
	return &backupRepository{db: db}
}

func (r *backupRepository) Create(ctx context.Context, backup *model.Backup) error {
	return r.db.WithContext(ctx).Create(backup).Error
}

func (r *backupRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Backup, error) {
	var backup model.Backup
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&backup).Error; err != nil {
		return nil, err
	}
	return &backup, nil
}

func (r *backupRepository) Update(ctx context.Context, backup *model.Backup) error {
	return r.db.WithContext(ctx).Save(backup).Error
}

func (r *backupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Backup{}, "id = ?", id).Error
}

func (r *backupRepository) ListByService(ctx context.Context, serviceID uuid.UUID, filter *model.ListBackupsRequest) ([]*model.Backup, int64, error) {
	var backups []*model.Backup
	query := r.db.WithContext(ctx).Model(&model.Backup{}).Where("service_id = ?", serviceID)

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	total, paginated, err := pagination.Paginate(query, filter.Page, filter.PageSize)
	if err != nil {
		return nil, 0, err
	}

	if err := paginated.Order("last_backup_time DESC").Find(&backups).Error; err != nil {
		return nil, 0, err
	}

	return backups, total, nil
}
