package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MonitoringConfigRepository interface {
	GetByService(ctx context.Context, serviceID uuid.UUID) (*model.MonitoringConfig, error)
	Upsert(ctx context.Context, config *model.MonitoringConfig) error
	FindEnabled(ctx context.Context, configs *[]model.MonitoringConfig) error
}

type monitoringConfigRepository struct {
	db *gorm.DB
}

func NewMonitoringConfigRepository(db *gorm.DB) MonitoringConfigRepository {
	return &monitoringConfigRepository{db: db}
}

func (r *monitoringConfigRepository) GetByService(ctx context.Context, serviceID uuid.UUID) (*model.MonitoringConfig, error) {
	var config model.MonitoringConfig
	if err := r.db.WithContext(ctx).Where("service_id = ?", serviceID).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *monitoringConfigRepository) Upsert(ctx context.Context, config *model.MonitoringConfig) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "service_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"enabled", "ping_interval_seconds", "timeout_seconds", "retries"}),
	}).Create(config).Error
}

func (r *monitoringConfigRepository) FindEnabled(ctx context.Context, configs *[]model.MonitoringConfig) error {
	return r.db.WithContext(ctx).Where("enabled = ?", true).Find(configs).Error
}
