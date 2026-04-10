package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/pkg/pagination"
	"gorm.io/gorm"
)

type DeploymentRepository interface {
	Create(ctx context.Context, deployment *model.Deployment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Deployment, error)
	Update(ctx context.Context, deployment *model.Deployment) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByService(ctx context.Context, serviceID uuid.UUID, filter *model.ListDeploymentsRequest) ([]*model.Deployment, int64, error)
}

type deploymentRepository struct {
	db *gorm.DB
}

func NewDeploymentRepository(db *gorm.DB) DeploymentRepository {
	return &deploymentRepository{db: db}
}

func (r *deploymentRepository) Create(ctx context.Context, deployment *model.Deployment) error {
	return r.db.WithContext(ctx).Create(deployment).Error
}

func (r *deploymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Deployment, error) {
	var deployment model.Deployment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&deployment).Error; err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (r *deploymentRepository) Update(ctx context.Context, deployment *model.Deployment) error {
	return r.db.WithContext(ctx).Save(deployment).Error
}

func (r *deploymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Deployment{}, "id = ?", id).Error
}

func (r *deploymentRepository) ListByService(ctx context.Context, serviceID uuid.UUID, filter *model.ListDeploymentsRequest) ([]*model.Deployment, int64, error) {
	var deployments []*model.Deployment
	query := r.db.WithContext(ctx).Model(&model.Deployment{}).Where("service_id = ?", serviceID)

	total, paginated, err := pagination.Paginate(query, filter.Page, filter.PageSize)
	if err != nil {
		return nil, 0, err
	}

	if err := paginated.Order("created_at DESC").Find(&deployments).Error; err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}
