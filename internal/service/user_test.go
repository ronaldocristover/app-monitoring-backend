package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ronaldocristover/app-monitoring/internal/model"
)

// --- Mock UserRepository (shared with auth_test.go) ---

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, filter *model.ListUsersRequest) ([]*model.User, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.User), args.Get(1).(int64), args.Error(2)
}

// --- UserService helpers ---

func newTestUserService() (UserService, *MockUserRepository) {
	mockRepo := new(MockUserRepository)
	return NewUserService(mockRepo), mockRepo
}

// --- Tests ---

func TestUserService_GetByID_Success(t *testing.T) {
	svc, mockRepo := newTestUserService()
	userID := uuid.New()
	expected := &model.User{ID: userID, Email: "test@example.com", Name: "Test User"}

	mockRepo.On("GetByID", mock.Anything, userID).Return(expected, nil)

	user, err := svc.GetByID(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, expected, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	svc, mockRepo := newTestUserService()
	userID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)

	user, err := svc.GetByID(context.Background(), userID)

	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Update_Success(t *testing.T) {
	svc, mockRepo := newTestUserService()
	userID := uuid.New()
	existing := &model.User{ID: userID, Email: "old@example.com", Name: "Old Name"}

	mockRepo.On("GetByID", mock.Anything, userID).Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	user, err := svc.Update(context.Background(), userID, &model.UpdateUserRequest{
		Name: "New Name",
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", user.Name)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Update_NotFound(t *testing.T) {
	svc, mockRepo := newTestUserService()
	userID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)

	user, err := svc.Update(context.Background(), userID, &model.UpdateUserRequest{Name: "New"})

	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Update_DuplicateEmail(t *testing.T) {
	svc, mockRepo := newTestUserService()
	userID := uuid.New()
	otherID := uuid.New()
	existing := &model.User{ID: userID, Email: "old@example.com", Name: "Old Name"}
	duplicate := &model.User{ID: otherID, Email: "new@example.com"}

	mockRepo.On("GetByID", mock.Anything, userID).Return(existing, nil)
	mockRepo.On("GetByEmail", mock.Anything, "new@example.com").Return(duplicate, nil)

	user, err := svc.Update(context.Background(), userID, &model.UpdateUserRequest{
		Email: "new@example.com",
	})

	assert.Equal(t, ErrUserExists, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Delete_Success(t *testing.T) {
	svc, mockRepo := newTestUserService()
	userID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, userID).Return(&model.User{ID: userID}, nil)
	mockRepo.On("Delete", mock.Anything, userID).Return(nil)

	err := svc.Delete(context.Background(), userID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Delete_NotFound(t *testing.T) {
	svc, mockRepo := newTestUserService()
	userID := uuid.New()

	mockRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)

	err := svc.Delete(context.Background(), userID)

	assert.Equal(t, ErrUserNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_List_Success(t *testing.T) {
	svc, mockRepo := newTestUserService()
	users := []*model.User{
		{ID: uuid.New(), Email: "user1@example.com", Name: "User One"},
		{ID: uuid.New(), Email: "user2@example.com", Name: "User Two"},
	}

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListUsersRequest")).Return(users, int64(2), nil)

	result, total, err := svc.List(context.Background(), &model.ListUsersRequest{Page: 1, PageSize: 20})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestUserService_List_DefaultPagination(t *testing.T) {
	svc, mockRepo := newTestUserService()

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListUsersRequest")).Return([]*model.User{}, int64(0), nil)

	result, total, err := svc.List(context.Background(), &model.ListUsersRequest{})

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestUserService_List_Error(t *testing.T) {
	svc, mockRepo := newTestUserService()

	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*model.ListUsersRequest")).Return(nil, int64(0), assert.AnError)

	result, total, err := svc.List(context.Background(), &model.ListUsersRequest{Page: 1})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}
