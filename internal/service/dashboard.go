package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DashboardService interface {
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
}

type DashboardStats struct {
	TotalApps            int64                     `json:"total_apps"`
	TotalServices        int64                     `json:"total_services"`
	ServicesUp           int64                     `json:"services_up"`
	ServicesDown         int64                     `json:"services_down"`
	RecentIncidents      []RecentIncident          `json:"recent_incidents"`
	EnvironmentBreakdown []EnvironmentBreakdownItem `json:"environment_breakdown"`
}

type RecentIncident struct {
	ServiceID    uuid.UUID `json:"service_id"`
	ServiceName  string    `json:"service_name"`
	Status       string    `json:"status"`
	CheckedAt    time.Time `json:"checked_at"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

type EnvironmentBreakdownItem struct {
	AppName       string `json:"app_name"`
	EnvironmentID string `json:"environment_id"`
	Environment   string `json:"environment"`
	TotalServices int64  `json:"total_services"`
}

type dashboardService struct {
	appRepo           repository.AppRepository
	serviceRepo       repository.ServiceRepository
	monitoringLogRepo repository.MonitoringLogRepository
	environmentRepo   repository.EnvironmentRepository
	db                *gorm.DB
	logger            *zap.SugaredLogger
}

func NewDashboardService(
	appRepo repository.AppRepository,
	serviceRepo repository.ServiceRepository,
	monitoringLogRepo repository.MonitoringLogRepository,
	environmentRepo repository.EnvironmentRepository,
	db *gorm.DB,
	logger *zap.SugaredLogger,
) DashboardService {
	return &dashboardService{
		appRepo:           appRepo,
		serviceRepo:       serviceRepo,
		monitoringLogRepo: monitoringLogRepo,
		environmentRepo:   environmentRepo,
		db:                db,
		logger:            logger,
	}
}

func (s *dashboardService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	if err := s.db.WithContext(ctx).Model(&model.App{}).Count(&stats.TotalApps).Error; err != nil {
		s.logger.Errorw("failed to count apps", "error", err)
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&model.Service{}).Count(&stats.TotalServices).Error; err != nil {
		s.logger.Errorw("failed to count services", "error", err)
		return nil, err
	}

	type statusCount struct {
		Status string
		Count  int64
	}
	var counts []statusCount
	if err := s.db.WithContext(ctx).Model(&model.MonitoringLog{}).
		Select("status, count(*) as count").
		Where("checked_at = (SELECT MAX(ml2.checked_at) FROM monitoring_logs ml2 WHERE ml2.service_id = monitoring_logs.service_id)").
		Group("status").
		Find(&counts).Error; err != nil {
		s.logger.Errorw("failed to count service statuses", "error", err)
		return nil, err
	}

	for _, c := range counts {
		switch c.Status {
		case "up":
			stats.ServicesUp = c.Count
		case "down":
			stats.ServicesDown = c.Count
		}
	}

	var recentLogs []model.MonitoringLog
	if err := s.db.WithContext(ctx).
		Preload("Service").
		Where("status = ?", "down").
		Order("checked_at DESC").
		Limit(10).
		Find(&recentLogs).Error; err != nil {
		s.logger.Errorw("failed to fetch recent incidents", "error", err)
		return nil, err
	}

	incidents := make([]RecentIncident, 0, len(recentLogs))
	for _, lg := range recentLogs {
		name := ""
		if lg.Service != nil {
			name = lg.Service.Name
		}
		incidents = append(incidents, RecentIncident{
			ServiceID:    lg.ServiceID,
			ServiceName:  name,
			Status:       lg.Status,
			CheckedAt:    lg.CheckedAt,
			ErrorMessage: lg.ErrorMessage,
		})
	}
	stats.RecentIncidents = incidents

	type envBreakdown struct {
		AppName       string
		EnvironmentID uuid.UUID
		Environment   string
		TotalServices int64
	}
	var breakdown []envBreakdown
	if err := s.db.WithContext(ctx).
		Table("environments").
		Select("apps.app_name, environments.id as environment_id, environments.name as environment, count(services.id) as total_services").
		Joins("JOIN apps ON apps.id = environments.app_id").
		Joins("LEFT JOIN services ON services.environment_id = environments.id").
		Group("apps.app_name, environments.id, environments.name").
		Order("apps.app_name, environments.name").
		Find(&breakdown).Error; err != nil {
		s.logger.Errorw("failed to fetch environment breakdown", "error", err)
		return nil, err
	}

	items := make([]EnvironmentBreakdownItem, 0, len(breakdown))
	for _, b := range breakdown {
		items = append(items, EnvironmentBreakdownItem{
			AppName:       b.AppName,
			EnvironmentID: b.EnvironmentID.String(),
			Environment:   b.Environment,
			TotalServices: b.TotalServices,
		})
	}
	stats.EnvironmentBreakdown = items

	return stats, nil
}
