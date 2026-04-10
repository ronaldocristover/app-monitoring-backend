package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/pkg/apierror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBackupCreate_Success(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	existingService := &model.Service{ID: serviceID}
	req := &model.CreateBackupRequest{
		ServiceID: serviceID,
		Enabled:   true,
		Path:      "/backups/db",
		Schedule:  "0 2 * * *",
	}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(b *model.Backup) bool {
		return b.ServiceID == serviceID && b.Enabled && b.Path == "/backups/db" && b.Schedule == "0 2 * * *"
	})).Return(nil)

	result, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, serviceID, result.ServiceID)
	assert.True(t, result.Enabled)
	assert.Equal(t, "/backups/db", result.Path)
	assert.Equal(t, "0 2 * * *", result.Schedule)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestBackupCreate_ServiceNotFound(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	req := &model.CreateBackupRequest{ServiceID: serviceID}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(nil, errors.New("not found"))

	result, err := svc.Create(context.Background(), req)

	assert.Nil(t, result)
	assert.Equal(t, ErrServiceNotFound, err)
	svcRepo.AssertExpectations(t)
}

func TestBackupCreate_RepoError(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	existingService := &model.Service{ID: serviceID}
	req := &model.CreateBackupRequest{ServiceID: serviceID}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	result, err := svc.Create(context.Background(), req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestBackupGetByID_Success(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	expected := &model.Backup{ID: id, Path: "/backups/db"}

	repo.On("GetByID", mock.Anything, id).Return(expected, nil)

	result, err := svc.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestBackupGetByID_NotFound(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	result, err := svc.GetByID(context.Background(), id)

	assert.Nil(t, result)
	assert.Equal(t, ErrBackupNotFound, err)
	repo.AssertExpectations(t)
}

func TestBackupUpdate_Success(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Backup{ID: id, Path: "/old", Schedule: "0 0 * * *", Enabled: false}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(b *model.Backup) bool {
		return b.Path == "/new" && b.Schedule == "0 3 * * *" && b.Enabled && b.Status == "active"
	})).Return(nil)

	enabled := true
	req := &model.UpdateBackupRequest{
		Enabled:  &enabled,
		Path:     "/new",
		Schedule: "0 3 * * *",
		Status:   "active",
	}
	result, err := svc.Update(context.Background(), id, req)

	assert.NoError(t, err)
	assert.Equal(t, "/new", result.Path)
	assert.Equal(t, "0 3 * * *", result.Schedule)
	assert.True(t, result.Enabled)
	assert.Equal(t, "active", result.Status)
	repo.AssertExpectations(t)
}

func TestBackupUpdate_NotFound(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	req := &model.UpdateBackupRequest{Path: "/new"}
	result, err := svc.Update(context.Background(), id, req)

	assert.Nil(t, result)
	assert.Equal(t, ErrBackupNotFound, err)
	repo.AssertExpectations(t)
}

func TestBackupUpdate_RepoError(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Backup{ID: id, Path: "/old"}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))

	req := &model.UpdateBackupRequest{Path: "/new"}
	result, err := svc.Update(context.Background(), id, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	repo.AssertExpectations(t)
}

func TestBackupDelete_Success(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Backup{ID: id}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Delete", mock.Anything, id).Return(nil)

	err := svc.Delete(context.Background(), id)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestBackupDelete_NotFound(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	err := svc.Delete(context.Background(), id)

	assert.Equal(t, ErrBackupNotFound, err)
	repo.AssertExpectations(t)
}

func TestBackupDelete_RepoError(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Backup{ID: id}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Delete", mock.Anything, id).Return(errors.New("db error"))

	err := svc.Delete(context.Background(), id)

	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	repo.AssertExpectations(t)
}

func TestBackupListByService_Success(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	existingService := &model.Service{ID: serviceID}
	backups := []*model.Backup{
		{ID: uuid.New(), ServiceID: serviceID, Path: "/backup1"},
		{ID: uuid.New(), ServiceID: serviceID, Path: "/backup2"},
	}

	svcRepo.On("GetByID", mock.Anything, serviceID).Return(existingService, nil)
	repo.On("ListByService", mock.Anything, serviceID, mock.Anything).Return(backups, int64(2), nil)

	result, total, err := svc.ListByService(context.Background(), serviceID, &model.ListBackupsRequest{})

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	svcRepo.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestBackupListByService_ServiceNotFound(t *testing.T) {
	repo := new(MockBackupRepository)
	svcRepo := new(MockServiceRepository)
	svc := NewBackupService(repo, svcRepo, newTestLogger())

	serviceID := uuid.New()
	svcRepo.On("GetByID", mock.Anything, serviceID).Return(nil, errors.New("not found"))

	result, total, err := svc.ListByService(context.Background(), serviceID, &model.ListBackupsRequest{})

	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, ErrServiceNotFound, err)
	svcRepo.AssertExpectations(t)
}
