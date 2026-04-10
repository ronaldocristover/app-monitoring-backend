package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/pkg/pagination"
	"gorm.io/gorm"
)

type EnvironmentRepository interface {
	Create(ctx context.Context, env *model.Environment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Environment, error)
	Update(ctx context.Context, env *model.Environment) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *model.ListEnvironmentsRequest) ([]*model.Environment, int64, error)
}

type environmentRepository struct {
	db *gorm.DB
}

func NewEnvironmentRepository(db *gorm.DB) EnvironmentRepository {
	return &environmentRepository{db: db}
}

func (r *environmentRepository) Create(ctx context.Context, env *model.Environment) error {
	return r.db.WithContext(ctx).Create(env).Error
}

func (r *environmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Environment, error) {
	var env model.Environment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&env).Error; err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *environmentRepository) Update(ctx context.Context, env *model.Environment) error {
	return r.db.WithContext(ctx).Save(env).Error
}

func (r *environmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("environment_id = ?", id).Delete(&model.Service{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Environment{}, "id = ?", id).Error
	})
}

func (r *environmentRepository) List(ctx context.Context, filter *model.ListEnvironmentsRequest) ([]*model.Environment, int64, error) {
	var envs []*model.Environment
	query := r.db.WithContext(ctx).Model(&model.Environment{})

	if filter.AppID != "" {
		if appID, err := uuid.Parse(filter.AppID); err == nil {
			query = query.Where("app_id = ?", appID)
		}
	}

	total, paginated, err := pagination.Paginate(query, filter.Page, filter.PageSize)
	if err != nil {
		return nil, 0, err
	}

	if err := paginated.Find(&envs).Error; err != nil {
		return nil, 0, err
	}

	return envs, total, nil
}
