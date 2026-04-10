package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/pkg/pagination"
	"gorm.io/gorm"
)

type ServerRepository interface {
	Create(ctx context.Context, server *model.Server) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Server, error)
	Update(ctx context.Context, server *model.Server) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *model.ListServersRequest) ([]*model.Server, int64, error)
}

type serverRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) ServerRepository {
	return &serverRepository{db: db}
}

func (r *serverRepository) Create(ctx context.Context, server *model.Server) error {
	return r.db.WithContext(ctx).Create(server).Error
}

func (r *serverRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Server, error) {
	var server model.Server
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&server).Error; err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *serverRepository) Update(ctx context.Context, server *model.Server) error {
	return r.db.WithContext(ctx).Save(server).Error
}

func (r *serverRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Server{}, "id = ?", id).Error
}

func (r *serverRepository) List(ctx context.Context, filter *model.ListServersRequest) ([]*model.Server, int64, error) {
	var servers []*model.Server
	query := r.db.WithContext(ctx).Model(&model.Server{})

	if filter.Search != "" {
		search := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("name ILIKE ? OR ip ILIKE ?", search, search)
	}

	total, paginated, err := pagination.Paginate(query, filter.Page, filter.PageSize)
	if err != nil {
		return nil, 0, err
	}

	if err := paginated.Find(&servers).Error; err != nil {
		return nil, 0, err
	}

	return servers, total, nil
}
