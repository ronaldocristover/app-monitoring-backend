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

func newTestAppService() (AppService, *MockAppRepository) {
	mockRepo := new(MockAppRepository)
	logger := zap.NewNop().Sugar()
	// Pass nil for db — CreateFull/UpdateFull require a real DB
	return NewAppService(mockRepo, nil, logger), mockRepo
}

// --- Tests ---

func TestAppService_Create_Success(t *testing.T) {
	svc, mockRepo := newTestAppService()
	req := &model.CreateAppRequest{
		AppName:     "My App",
		Description: "A test app",
		Tags:        "go,backend",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.App")).Return(nil)

	app, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, req.AppName, app.AppName)
	assert.Equal(t, req.Description, app.Description)
	assert.Equal(t, req.Tags, app.Tags)
	mockRepo.AssertExpectations(t)
}

func TestAppService_Create_RepoError(t *testing.T) {
	svc, mockRepo := newTestAppService()
	req := &model.CreateAppRequest{AppName: "My App"}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.App")).Return(assert.AnError)

	app, err := svc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, app)
	mockRepo.AssertExpectations(t)
}

func TestAppService_GetByID_Success(t *testing.T) {
	svc, mockRepo := newTestAppService()
	appID := uuid.New()
	expected := &model.App{ID: appID, AppName: "My App"}

	mockRepo.On("GetByID", mock.Anything, appID).Return(expected, nil)

	app, err := svc.GetByID(context.Background(), appID)

	assert.NoError(t, err)
	assert.Equal(t, expected, app)
	mockRepo.AssertExpectations(t)
}

func TestAppService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo := newTestAppService()
	appID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, appID).Return(nil, assert.AnError)

	app, err := svc.GetByID(context.Background(), appID)

	assert.Equal(t, ErrAppNotFound, err)
	assert.Nil(t, app)
	mockRepo.AssertExpectations(t)
}

func TestAppService_GetByIDFull_Success(t *testing.T) {
	svc, mockRepo := newTestAppService()
	appID := uuid.New()
	expected := &model.App{
		ID:      appID,
		AppName: "My App",
		Environments: []model.Environment{
			{ID: uuid.New(), AppID: appID, Name: "production"},
		},
	}

	mockRepo.On("GetByIDFull", mock.Anything, appID).Return(expected, nil)

	app, err := svc.GetByIDFull(context.Background(), appID)

	assert.NoError(t, err)
	assert.Equal(t, expected, app)
	assert.Len(t, app.Environments, 1)
	mockRepo.AssertExpectations(t)
}

func TestAppService_GetByIDFull_NotFound(t *testing.T) {
	svc, mockRepo := newTestAppService()
	appID := uuid.New()

	mockRepo.On("GetByIDFull", mock.Anything, appID).Return(nil, assert.AnError)

	app, err := svc.GetByIDFull(context.Background(), appID)

	assert.Equal(t, ErrAppNotFound, err)
	assert.Nil(t, app)
	mockRepo.AssertExpectations(t)
}

func TestAppService_Update_Success(t *testing.T) {
	svc, mockRepo := newTestAppService()
	appID := uuid.New()
	existing := &model.App{ID: appID, AppName: "Old Name", Description: "Old Desc"}
	updated := &model.App{ID: appID, AppName: "New Name", Description: "New Desc"}

	mockRepo.On("GetByID", mock.Anything, appID).Return(existing, nil).Once()
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.App")).Return(nil)
	mockRepo.On("GetByID", mock.Anything, appID).Return(updated, nil).Once()

	app, err := svc.Update(context.Background(), appID, &model.UpdateAppRequest{
		AppName:     "New Name",
		Description: "New Desc",
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", app.AppName)
	mockRepo.AssertExpectations(t)
}

func TestAppService_Update_NotFound(t *testing.T) {
	svc, mockRepo := newTestAppService()
	appID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, appID).Return(nil, assert.AnError)

	app, err := svc.Update(context.Background(), appID, &model.UpdateAppRequest{AppName: "New"})

	assert.Equal(t, ErrAppNotFound, err)
	assert.Nil(t, app)
	mockRepo.AssertExpectations(t)
}

func TestAppService_Update_RepoError(t *testing.T) {
	svc, mockRepo := newTestAppService()
	appID := uuid.New()
	existing := &model.App{ID: appID, AppName: "Old Name"}

	mockRepo.On("GetByID", mock.Anything, appID).Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.App")).Return(assert.AnError)

	app, err := svc.Update(context.Background(), appID, &model.UpdateAppRequest{AppName: "New"})

	assert.Error(t, err)
	assert.Nil(t, app)
	mockRepo.AssertExpectations(t)
}

func TestAppService_Delete_Success(t *testing.T) {
	svc, mockRepo := newTestAppService()
	appID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, appID).Return(&model.App{ID: appID}, nil)
	mockRepo.On("Delete", mock.Anything, appID).Return(nil)

	err := svc.Delete(context.Background(), appID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAppService_Delete_NotFound(t *testing.T) {
	svc, mockRepo := newTestAppService()
	appID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, appID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), appID)

	assert.Equal(t, ErrAppNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestAppService_Delete_RepoError(t *testing.T) {
	svc, mockRepo := newTestAppService()
	appID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, appID).Return(&model.App{ID: appID}, nil)
	mockRepo.On("Delete", mock.Anything, appID).Return(assert.AnError)

	err := svc.Delete(context.Background(), appID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAppService_List_Success(t *testing.T) {
	svc, mockRepo := newTestAppService()
	apps := []*model.App{
		{ID: uuid.New(), AppName: "App 1"},
		{ID: uuid.New(), AppName: "App 2"},
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListAppsRequest")).Return(apps, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListAppsRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestAppService_List_Empty(t *testing.T) {
	svc, mockRepo := newTestAppService()

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListAppsRequest")).Return([]*model.App{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListAppsRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}
