package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/internal/repository"
	"github.com/ronaldocristover/app-monitoring/pkg/apierror"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrAppNotFound = errors.New("app not found")
)

type AppService interface {
	Create(ctx context.Context, req *model.CreateAppRequest) (*model.App, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.App, error)
	GetByIDFull(ctx context.Context, id uuid.UUID) (*model.App, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateAppRequest) (*model.App, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *model.ListAppsRequest) ([]*model.App, int64, error)
	CreateFull(ctx context.Context, req *model.CreateFullAppRequest) (*model.App, error)
	UpdateFull(ctx context.Context, id uuid.UUID, req *model.UpdateFullAppRequest) (*model.App, error)
}

type appService struct {
	repo   repository.AppRepository
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewAppService(repo repository.AppRepository, db *gorm.DB, logger *zap.SugaredLogger) AppService {
	return &appService{repo: repo, db: db, logger: logger}
}

func (s *appService) Create(ctx context.Context, req *model.CreateAppRequest) (*model.App, error) {
	app := &model.App{
		AppName:     req.AppName,
		Description: req.Description,
		Tags:        req.Tags,
	}
	if err := s.repo.Create(ctx, app); err != nil {
		s.logger.Errorw("failed to create app", "error", err)
		return nil, apierror.Internal("Failed to create app")
	}
	return app, nil
}

func (s *appService) GetByID(ctx context.Context, id uuid.UUID) (*model.App, error) {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrAppNotFound
	}
	return app, nil
}

func (s *appService) GetByIDFull(ctx context.Context, id uuid.UUID) (*model.App, error) {
	app, err := s.repo.GetByIDFull(ctx, id)
	if err != nil {
		return nil, ErrAppNotFound
	}
	return app, nil
}

func (s *appService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateAppRequest) (*model.App, error) {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrAppNotFound
	}
	if req.AppName != "" {
		app.AppName = req.AppName
	}
	if req.Description != "" {
		app.Description = req.Description
	}
	if req.Tags != "" {
		app.Tags = req.Tags
	}
	if err := s.repo.Update(ctx, app); err != nil {
		s.logger.Errorw("failed to update app", "error", err)
		return nil, apierror.Internal("Failed to update app")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *appService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrAppNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Errorw("failed to delete app", "error", err)
		return apierror.Internal("Failed to delete app")
	}
	return nil
}

func (s *appService) List(ctx context.Context, req *model.ListAppsRequest) ([]*model.App, int64, error) {
	return s.repo.List(ctx, req)
}

func (s *appService) CreateFull(ctx context.Context, req *model.CreateFullAppRequest) (*model.App, error) {
	app := &model.App{
		AppName:     req.AppName,
		Description: req.Description,
		Tags:        req.Tags,
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(app).Error; err != nil {
			return err
		}
		for _, envInput := range req.Environments {
			env := &model.Environment{
				AppID: app.ID,
				Name:  envInput.Name,
			}
			if err := tx.Create(env).Error; err != nil {
				return err
			}
			for _, svcInput := range envInput.Services {
				svc := &model.Service{
					EnvironmentID:  env.ID,
					ServerID:       svcInput.ServerID,
					Name:           svcInput.Name,
					Type:           svcInput.Type,
					URL:            svcInput.URL,
					Repository:     svcInput.Repository,
					StackLanguage:  svcInput.StackLanguage,
					StackFramework: svcInput.StackFramework,
					DBType:         svcInput.DBType,
					DBHost:         svcInput.DBHost,
				}
				if err := tx.Create(svc).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		s.logger.Errorw("failed to create full app", "error", err)
		return nil, apierror.Internal("Failed to create app with nested resources")
	}
	return s.repo.GetByIDFull(ctx, app.ID)
}

func (s *appService) UpdateFull(ctx context.Context, id uuid.UUID, req *model.UpdateFullAppRequest) (*model.App, error) {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrAppNotFound
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.AppName != "" {
			app.AppName = req.AppName
		}
		if req.Description != "" {
			app.Description = req.Description
		}
		if req.Tags != "" {
			app.Tags = req.Tags
		}
		if err := tx.Save(app).Error; err != nil {
			return err
		}

		for _, envInput := range req.Environments {
			var env model.Environment
			if envInput.ID != nil {
				// Update existing environment
				if err := tx.Where("id = ? AND app_id = ?", *envInput.ID, id).First(&env).Error; err != nil {
					return err
				}
				env.Name = envInput.Name
				if err := tx.Save(&env).Error; err != nil {
					return err
				}
			} else {
				// Create new environment
				env = model.Environment{
					AppID: app.ID,
					Name:  envInput.Name,
				}
				if err := tx.Create(&env).Error; err != nil {
					return err
				}
			}

			for _, svcInput := range envInput.Services {
				if svcInput.ID != nil {
					// Update existing service
					var svc model.Service
					if err := tx.Where("id = ? AND environment_id = ?", *svcInput.ID, env.ID).First(&svc).Error; err != nil {
						return err
					}
					svc.ServerID = svcInput.ServerID
					svc.Name = svcInput.Name
					svc.Type = svcInput.Type
					svc.URL = svcInput.URL
					svc.Repository = svcInput.Repository
					svc.StackLanguage = svcInput.StackLanguage
					svc.StackFramework = svcInput.StackFramework
					svc.DBType = svcInput.DBType
					svc.DBHost = svcInput.DBHost
					if err := tx.Save(&svc).Error; err != nil {
						return err
					}
				} else {
					// Create new service
					svc := model.Service{
						EnvironmentID:  env.ID,
						ServerID:       svcInput.ServerID,
						Name:           svcInput.Name,
						Type:           svcInput.Type,
						URL:            svcInput.URL,
						Repository:     svcInput.Repository,
						StackLanguage:  svcInput.StackLanguage,
						StackFramework: svcInput.StackFramework,
						DBType:         svcInput.DBType,
						DBHost:         svcInput.DBHost,
					}
					if err := tx.Create(&svc).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		s.logger.Errorw("failed to update full app", "error", err)
		return nil, apierror.Internal("Failed to update app with nested resources")
	}
	return s.repo.GetByIDFull(ctx, id)
}
