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

// --- Mock ServerRepository ---

type MockServerRepository struct {
	mock.Mock
}

func (m *MockServerRepository) Create(ctx context.Context, server *model.Server) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockServerRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Server, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Server), args.Error(1)
}

func (m *MockServerRepository) Update(ctx context.Context, server *model.Server) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockServerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServerRepository) List(ctx context.Context, filter *model.ListServersRequest) ([]*model.Server, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.Server), args.Get(1).(int64), args.Error(2)
}

// --- Helpers ---

func newTestServerService() (ServerService, *MockServerRepository) {
	mockRepo := new(MockServerRepository)
	logger := zap.NewNop().Sugar()
	return NewServerService(mockRepo, logger), mockRepo
}

// --- Tests ---

func TestServerService_Create_Success(t *testing.T) {
	svc, mockRepo := newTestServerService()
	req := &model.CreateServerRequest{
		Name:     "prod-server-1",
		IP:       "10.0.0.1",
		Provider: "aws",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Server")).Return(nil)

	server, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, "prod-server-1", server.Name)
	assert.Equal(t, "10.0.0.1", server.IP)
	assert.Equal(t, "aws", server.Provider)
	mockRepo.AssertExpectations(t)
}

func TestServerService_Create_RepoError(t *testing.T) {
	svc, mockRepo := newTestServerService()
	req := &model.CreateServerRequest{
		Name: "prod-server-1",
		IP:   "10.0.0.1",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Server")).Return(assert.AnError)

	server, err := svc.Create(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, server)
	mockRepo.AssertExpectations(t)
}

func TestServerService_GetByID_Success(t *testing.T) {
	svc, mockRepo := newTestServerService()
	serverID := uuid.New()
	expected := &model.Server{ID: serverID, Name: "prod-server-1", IP: "10.0.0.1"}

	mockRepo.On("GetByID", mock.Anything, serverID).Return(expected, nil)

	server, err := svc.GetByID(context.Background(), serverID)

	assert.NoError(t, err)
	assert.Equal(t, expected, server)
	mockRepo.AssertExpectations(t)
}

func TestServerService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo := newTestServerService()
	serverID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, serverID).Return(nil, assert.AnError)

	server, err := svc.GetByID(context.Background(), serverID)

	assert.Equal(t, ErrServerNotFound, err)
	assert.Nil(t, server)
	mockRepo.AssertExpectations(t)
}

func TestServerService_Update_Success(t *testing.T) {
	svc, mockRepo := newTestServerService()
	serverID := uuid.New()
	existing := &model.Server{ID: serverID, Name: "old-name", IP: "10.0.0.1"}
	updated := &model.Server{ID: serverID, Name: "new-name", IP: "10.0.0.2", Provider: "gcp"}

	mockRepo.On("GetByID", mock.Anything, serverID).Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Server")).Return(nil)
	mockRepo.On("GetByID", mock.Anything, serverID).Return(updated, nil)

	server, err := svc.Update(context.Background(), serverID, &model.UpdateServerRequest{
		Name:     "new-name",
		IP:       "10.0.0.2",
		Provider: "gcp",
	})

	assert.NoError(t, err)
	assert.Equal(t, "new-name", server.Name)
	assert.Equal(t, "10.0.0.2", server.IP)
	assert.Equal(t, "gcp", server.Provider)
	mockRepo.AssertExpectations(t)
}

func TestServerService_Update_NotFound(t *testing.T) {
	svc, mockRepo := newTestServerService()
	serverID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, serverID).Return(nil, assert.AnError)

	server, err := svc.Update(context.Background(), serverID, &model.UpdateServerRequest{Name: "new"})

	assert.Equal(t, ErrServerNotFound, err)
	assert.Nil(t, server)
	mockRepo.AssertExpectations(t)
}

func TestServerService_Update_RepoError(t *testing.T) {
	svc, mockRepo := newTestServerService()
	serverID := uuid.New()
	existing := &model.Server{ID: serverID, Name: "old-name", IP: "10.0.0.1"}

	mockRepo.On("GetByID", mock.Anything, serverID).Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.Server")).Return(assert.AnError)

	server, err := svc.Update(context.Background(), serverID, &model.UpdateServerRequest{Name: "new"})

	assert.Error(t, err)
	assert.Nil(t, server)
	mockRepo.AssertExpectations(t)
}

func TestServerService_Delete_Success(t *testing.T) {
	svc, mockRepo := newTestServerService()
	serverID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, serverID).Return(&model.Server{ID: serverID}, nil)
	mockRepo.On("Delete", mock.Anything, serverID).Return(nil)

	err := svc.Delete(context.Background(), serverID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServerService_Delete_NotFound(t *testing.T) {
	svc, mockRepo := newTestServerService()
	serverID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, serverID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), serverID)

	assert.Equal(t, ErrServerNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestServerService_Delete_RepoError(t *testing.T) {
	svc, mockRepo := newTestServerService()
	serverID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, serverID).Return(&model.Server{ID: serverID}, nil)
	mockRepo.On("Delete", mock.Anything, serverID).Return(assert.AnError)

	err := svc.Delete(context.Background(), serverID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestServerService_List_Success(t *testing.T) {
	svc, mockRepo := newTestServerService()
	servers := []*model.Server{
		{ID: uuid.New(), Name: "server-1", IP: "10.0.0.1"},
		{ID: uuid.New(), Name: "server-2", IP: "10.0.0.2"},
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListServersRequest")).Return(servers, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListServersRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestServerService_List_Empty(t *testing.T) {
	svc, mockRepo := newTestServerService()

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListServersRequest")).Return([]*model.Server{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListServersRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestServerService_List_Error(t *testing.T) {
	svc, mockRepo := newTestServerService()

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListServersRequest")).Return(nil, int64(0), assert.AnError)

	result, total, err := svc.List(context.Background(), &model.ListServersRequest{Page: 1})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}
