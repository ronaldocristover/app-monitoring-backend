package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/stretchr/testify/mock"
)

// MockServiceRepository mocks repository.ServiceRepository
type MockServiceRepository struct {
	mock.Mock
}

func (m *MockServiceRepository) Create(ctx context.Context, svc *model.Service) error {
	args := m.Called(ctx, svc)
	return args.Error(0)
}

func (m *MockServiceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Service, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*model.Service), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServiceRepository) GetByIDFull(ctx context.Context, id uuid.UUID) (*model.Service, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*model.Service), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockServiceRepository) Update(ctx context.Context, svc *model.Service) error {
	args := m.Called(ctx, svc)
	return args.Error(0)
}

func (m *MockServiceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServiceRepository) List(ctx context.Context, filter *model.ListServicesRequest) ([]*model.Service, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Service), args.Get(1).(int64), args.Error(2)
}

// MockMonitoringLogRepository mocks repository.MonitoringLogRepository
type MockMonitoringLogRepository struct {
	mock.Mock
}

func (m *MockMonitoringLogRepository) Create(ctx context.Context, log *model.MonitoringLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockMonitoringLogRepository) ListByService(ctx context.Context, serviceID uuid.UUID, filter *model.ListMonitoringLogsRequest) ([]*model.MonitoringLog, int64, error) {
	args := m.Called(ctx, serviceID, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.MonitoringLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockMonitoringLogRepository) GetLatest(ctx context.Context, serviceID uuid.UUID) (*model.MonitoringLog, error) {
	args := m.Called(ctx, serviceID)
	if v := args.Get(0); v != nil {
		return v.(*model.MonitoringLog), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockMonitoringConfigRepository mocks repository.MonitoringConfigRepository
type MockMonitoringConfigRepository struct {
	mock.Mock
}

func (m *MockMonitoringConfigRepository) GetByService(ctx context.Context, serviceID uuid.UUID) (*model.MonitoringConfig, error) {
	args := m.Called(ctx, serviceID)
	if v := args.Get(0); v != nil {
		return v.(*model.MonitoringConfig), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockMonitoringConfigRepository) Upsert(ctx context.Context, config *model.MonitoringConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockMonitoringConfigRepository) FindEnabled(ctx context.Context, configs *[]model.MonitoringConfig) error {
	args := m.Called(ctx, configs)
	return args.Error(0)
}

// MockDeploymentRepository mocks repository.DeploymentRepository
type MockDeploymentRepository struct {
	mock.Mock
}

func (m *MockDeploymentRepository) Create(ctx context.Context, d *model.Deployment) error {
	args := m.Called(ctx, d)
	return args.Error(0)
}

func (m *MockDeploymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Deployment, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*model.Deployment), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockDeploymentRepository) Update(ctx context.Context, d *model.Deployment) error {
	args := m.Called(ctx, d)
	return args.Error(0)
}

func (m *MockDeploymentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeploymentRepository) ListByService(ctx context.Context, serviceID uuid.UUID, filter *model.ListDeploymentsRequest) ([]*model.Deployment, int64, error) {
	args := m.Called(ctx, serviceID, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Deployment), args.Get(1).(int64), args.Error(2)
}

// MockBackupRepository mocks repository.BackupRepository
type MockBackupRepository struct {
	mock.Mock
}

func (m *MockBackupRepository) Create(ctx context.Context, b *model.Backup) error {
	args := m.Called(ctx, b)
	return args.Error(0)
}

func (m *MockBackupRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Backup, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*model.Backup), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockBackupRepository) Update(ctx context.Context, b *model.Backup) error {
	args := m.Called(ctx, b)
	return args.Error(0)
}

func (m *MockBackupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBackupRepository) ListByService(ctx context.Context, serviceID uuid.UUID, filter *model.ListBackupsRequest) ([]*model.Backup, int64, error) {
	args := m.Called(ctx, serviceID, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Backup), args.Get(1).(int64), args.Error(2)
}

// MockAppRepository mocks repository.AppRepository
type MockAppRepository struct {
	mock.Mock
}

func (m *MockAppRepository) Create(ctx context.Context, app *model.App) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockAppRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.App, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*model.App), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAppRepository) GetByIDFull(ctx context.Context, id uuid.UUID) (*model.App, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*model.App), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAppRepository) Update(ctx context.Context, app *model.App) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockAppRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAppRepository) List(ctx context.Context, filter *model.ListAppsRequest) ([]*model.App, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.App), args.Get(1).(int64), args.Error(2)
}

// MockEnvironmentRepository mocks repository.EnvironmentRepository
type MockEnvironmentRepository struct {
	mock.Mock
}

func (m *MockEnvironmentRepository) Create(ctx context.Context, env *model.Environment) error {
	args := m.Called(ctx, env)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Environment, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*model.Environment), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockEnvironmentRepository) Update(ctx context.Context, env *model.Environment) error {
	args := m.Called(ctx, env)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) List(ctx context.Context, filter *model.ListEnvironmentsRequest) ([]*model.Environment, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Environment), args.Get(1).(int64), args.Error(2)
}
