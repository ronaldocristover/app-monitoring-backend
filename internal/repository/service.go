package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/pkg/pagination"
	"gorm.io/gorm"
)

type ServiceRepository interface {
	Create(ctx context.Context, svc *model.Service) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Service, error)
	GetByIDFull(ctx context.Context, id uuid.UUID) (*model.Service, error)
	Update(ctx context.Context, svc *model.Service) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *model.ListServicesRequest) ([]*model.Service, int64, error)
}

type serviceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) ServiceRepository {
	return &serviceRepository{db: db}
}

func (r *serviceRepository) Create(ctx context.Context, svc *model.Service) error {
	return r.db.WithContext(ctx).Create(svc).Error
}

func (r *serviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Service, error) {
	var svc model.Service
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&svc).Error; err != nil {
		return nil, err
	}
	return &svc, nil
}

func (r *serviceRepository) GetByIDFull(ctx context.Context, id uuid.UUID) (*model.Service, error) {
	var svc model.Service
	if err := r.db.WithContext(ctx).
		Preload("Environment").
		Preload("Server").
		Preload("MonitoringConfig").
		Preload("Backups").
		Preload("Deployments").
		Where("id = ?", id).First(&svc).Error; err != nil {
		return nil, err
	}
	return &svc, nil
}

func (r *serviceRepository) Update(ctx context.Context, svc *model.Service) error {
	return r.db.WithContext(ctx).Save(svc).Error
}

func (r *serviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Service{}, "id = ?", id).Error
}

func (r *serviceRepository) List(ctx context.Context, filter *model.ListServicesRequest) ([]*model.Service, int64, error) {
	var services []*model.Service
	query := r.db.WithContext(ctx).Model(&model.Service{})

	if filter.EnvironmentID != "" {
		envID, err := uuid.Parse(filter.EnvironmentID)
		if err == nil {
			query = query.Where("environment_id = ?", envID)
		}
	}
	if filter.ServerID != "" {
		srvID, err := uuid.Parse(filter.ServerID)
		if err == nil {
			query = query.Where("server_id = ?", srvID)
		}
	}
	if filter.Search != "" {
		search := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("name ILIKE ?", search)
	}

	total, paginated, err := pagination.Paginate(query, filter.Page, filter.PageSize)
	if err != nil {
		return nil, 0, err
	}

	if err := paginated.Order("created_at DESC").Find(&services).Error; err != nil {
		return nil, 0, err
	}

	return services, total, nil
}
