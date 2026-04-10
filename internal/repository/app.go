package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/pkg/pagination"
	"gorm.io/gorm"
)

type AppRepository interface {
	Create(ctx context.Context, app *model.App) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.App, error)
	GetByIDFull(ctx context.Context, id uuid.UUID) (*model.App, error)
	Update(ctx context.Context, app *model.App) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *model.ListAppsRequest) ([]*model.App, int64, error)
}

type appRepository struct {
	db *gorm.DB
}

func NewAppRepository(db *gorm.DB) AppRepository {
	return &appRepository{db: db}
}

func (r *appRepository) Create(ctx context.Context, app *model.App) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *appRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.App, error) {
	var app model.App
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *appRepository) GetByIDFull(ctx context.Context, id uuid.UUID) (*model.App, error) {
	var app model.App
	if err := r.db.WithContext(ctx).
		Preload("Environments.Services").
		Where("id = ?", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *appRepository) Update(ctx context.Context, app *model.App) error {
	return r.db.WithContext(ctx).Save(app).Error
}

func (r *appRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete services belonging to environments of this app
		if err := tx.Where("environment_id IN (?)",
			tx.Model(&model.Environment{}).Select("id").Where("app_id = ?", id),
		).Delete(&model.Service{}).Error; err != nil {
			return err
		}
		// Delete environments
		if err := tx.Where("app_id = ?", id).Delete(&model.Environment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.App{}, "id = ?", id).Error
	})
}

func (r *appRepository) List(ctx context.Context, filter *model.ListAppsRequest) ([]*model.App, int64, error) {
	var apps []*model.App
	query := r.db.WithContext(ctx).Model(&model.App{})

	if filter.Search != "" {
		search := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("app_name ILIKE ?", search)
	}
	if filter.Tags != "" {
		tags := strings.Split(filter.Tags, ",")
		for _, tag := range tags {
			t := strings.TrimSpace(tag)
			if t != "" {
				query = query.Where("tags ILIKE ?", fmt.Sprintf("%%%s%%", t))
			}
		}
	}

	total, paginated, err := pagination.Paginate(query, filter.Page, filter.PageSize)
	if err != nil {
		return nil, 0, err
	}

	if err := paginated.Find(&apps).Error; err != nil {
		return nil, 0, err
	}

	// Populate total_services and total_environments for each app
	for _, app := range apps {
		r.db.WithContext(ctx).Model(&model.Environment{}).Where("app_id = ?", app.ID).Count(&app.TotalEnvironments)
		r.db.WithContext(ctx).
			Table("services").
			Joins("JOIN environments ON environments.id = services.environment_id").
			Where("environments.app_id = ?", app.ID).
			Count(&app.TotalServices)
	}

	return apps, total, nil
}
