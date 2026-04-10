package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/ronaldocristover/app-monitoring/internal/model"
)

// --- Helpers ---

func newTestEnvironmentService() (EnvironmentService, *MockEnvironmentRepository) {
	mockRepo := new(MockEnvironmentRepository)
	logger := zap.NewNop().Sugar()
	return NewEnvironmentService(mockRepo, logger), mockRepo
}

// --- Tests ---

func TestEnvironmentService_Create_Success(t *testing.T) {
	svc, mockRepo := newTestEnvironmentService()
	appID := uuid.New()
	req := &model.CreateEnvironmentRequest{
		AppID: appID,
		Name:  "production",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Environment")).Return(nil)

	env, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, env)
	assert.Equal(t, appID, env.AppID)
	assert.Equal(t, "production", env.Name)
	mockRepo.AssertExpectations(t)
}

func TestEnvironmentService_Create_RepoError(t *testing.T) {
	svc, mockRepo := newTestEnvironmentService()
	req := &model.CreateEnvironmentRequest{
		AppID: uuid.New(),
		Name:  "staging",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Environment")).Return(assert.AnError)

	env, err := svc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, env)
	mockRepo.AssertExpectations(t)
}

func TestEnvironmentService_GetByID_Success(t *testing.T) {
	svc, mockRepo := newTestEnvironmentService()
	envID := uuid.New()
	expected := &model.Environment{ID: envID, Name: "production"}

	mockRepo.On("GetByID", mock.Anything, envID).Return(expected, nil)

	env, err := svc.GetByID(context.Background(), envID)

	assert.NoError(t, err)
	assert.Equal(t, expected, env)
	mockRepo.AssertExpectations(t)
}

func TestEnvironmentService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo := newTestEnvironmentService()
	envID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, envID).Return(nil, assert.AnError)

	env, err := svc.GetByID(context.Background(), envID)

	assert.Equal(t, ErrEnvironmentNotFound, err)
	assert.Nil(t, env)
	mockRepo.AssertExpectations(t)
}

func TestEnvironmentService_Update_Success(t *testing.T) {
	svc, mockRepo := newTestEnvironmentService()
	envID := uuid.New()
	existing := &model.Environment{ID: envID, Name: "staging"}
	updated := &model.Environment{ID: envID, Name: "production"}

	mockRepo.On("GetByID", mock.Anything, envID).Return(existing, nil).Once()
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Environment")).Return(nil)
	mockRepo.On("GetByID", mock.Anything, envID).Return(updated, nil).Once()

	env, err := svc.Update(context.Background(), envID, &model.UpdateEnvironmentRequest{Name: "production"})

	assert.NoError(t, err)
	assert.Equal(t, "production", env.Name)
	mockRepo.AssertExpectations(t)
}

func TestEnvironmentService_Update_NotFound(t *testing.T) {
	svc, mockRepo := newTestEnvironmentService()
	envID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, envID).Return(nil, assert.AnError)

	env, err := svc.Update(context.Background(), envID, &model.UpdateEnvironmentRequest{Name: "new"})

	assert.Equal(t, ErrEnvironmentNotFound, err)
	assert.Nil(t, env)
	mockRepo.AssertExpectations(t)
}

func TestEnvironmentService_Update_RepoError(t *testing.T) {
	svc, mockRepo := newTestEnvironmentService()
	envID := uuid.New()
	existing := &model.Environment{ID: envID, Name: "staging"}

	mockRepo.On("GetByID", mock.Anything, envID).Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Environment")).Return(assert.AnError)

	env, err := svc.Update(context.Background(), envID, &model.UpdateEnvironmentRequest{Name: "production"})

	assert.Error(t, err)
	assert.Nil(t, env)
	mockRepo.AssertExpectations(t)
}

func TestEnvironmentService_Delete_Success(t *testing.T) {
	svc, mockRepo := newTestEnvironmentService()
	envID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, envID).Return(&model.Environment{ID: envID}, nil)
	mockRepo.On("Delete", mock.Anything, envID).Return(nil)

	err := svc.Delete(context.Background(), envID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestEnvironmentService_Delete_NotFound(t *testing.T) {
	svc, mockRepo := newTestEnvironmentService()
	envID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, envID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), envID)

	assert.Equal(t, ErrEnvironmentNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestEnvironmentService_Delete_RepoError(t *testing.T) {
	svc, mockRepo := newTestEnvironmentService()
	envID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, envID).Return(&model.Environment{ID: envID}, nil)
	mockRepo.On("Delete", mock.Anything, envID).Return(assert.AnError)

	err := svc.Delete(context.Background(), envID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestEnvironmentService_List_Success(t *testing.T) {
	svc, mockRepo := newTestEnvironmentService()
	envs := []*model.Environment{
		{ID: uuid.New(), Name: "production"},
		{ID: uuid.New(), Name: "staging"},
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListEnvironmentsRequest")).Return(envs, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListEnvironmentsRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestEnvironmentService_List_Empty(t *testing.T) {
	svc, mockRepo := newTestEnvironmentService()

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListEnvironmentsRequest")).Return([]*model.Environment{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListEnvironmentsRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}
