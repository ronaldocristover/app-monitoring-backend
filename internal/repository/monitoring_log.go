package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/pkg/pagination"
	"gorm.io/gorm"
)

type MonitoringLogRepository interface {
	Create(ctx context.Context, log *model.MonitoringLog) error
	ListByService(ctx context.Context, serviceID uuid.UUID, filter *model.ListMonitoringLogsRequest) ([]*model.MonitoringLog, int64, error)
	GetLatest(ctx context.Context, serviceID uuid.UUID) (*model.MonitoringLog, error)
}

type monitoringLogRepository struct {
	db *gorm.DB
}

func NewMonitoringLogRepository(db *gorm.DB) MonitoringLogRepository {
	return &monitoringLogRepository{db: db}
}

func (r *monitoringLogRepository) Create(ctx context.Context, log *model.MonitoringLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *monitoringLogRepository) ListByService(ctx context.Context, serviceID uuid.UUID, filter *model.ListMonitoringLogsRequest) ([]*model.MonitoringLog, int64, error) {
	var logs []*model.MonitoringLog
	query := r.db.WithContext(ctx).Model(&model.MonitoringLog{}).Where("service_id = ?", serviceID)

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	total, paginated, err := pagination.Paginate(query, filter.Page, filter.PageSize)
	if err != nil {
		return nil, 0, err
	}

	if err := paginated.Order("checked_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *monitoringLogRepository) GetLatest(ctx context.Context, serviceID uuid.UUID) (*model.MonitoringLog, error) {
	var log model.MonitoringLog
	if err := r.db.WithContext(ctx).Where("service_id = ?", serviceID).Order("checked_at DESC").First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}
