package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/ronaldocristover/app-monitoring/internal/model"
)

const testJWTSecret = "test-secret-key-at-least-32-characters-long"

func newTestAuthService() (AuthService, *MockUserRepository) {
	mockRepo := new(MockUserRepository)
	return NewAuthService(mockRepo, testJWTSecret, 15*time.Minute, 7*24*time.Hour), mockRepo
}

// --- Register ---

func TestAuthService_Register_Success(t *testing.T) {
	authSvc, mockRepo := newTestAuthService()
	req := &model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	resp, err := authSvc.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.Email, resp.User.Email)
	assert.Equal(t, req.Name, resp.User.Name)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	authSvc, mockRepo := newTestAuthService()
	req := &model.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	existingUser := &model.User{ID: uuid.New(), Email: req.Email}

	mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(existingUser, nil)

	resp, err := authSvc.Register(context.Background(), req)

	assert.Equal(t, ErrUserExists, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_CreateError(t *testing.T) {
	authSvc, mockRepo := newTestAuthService()
	req := &model.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	mockRepo.On("GetByEmail", mock.Anything, req.Email).Return(nil, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(assert.AnError)

	resp, err := authSvc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

// --- Login ---

func TestAuthService_Login_Success(t *testing.T) {
	authSvc, mockRepo := newTestAuthService()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Name:         "Test User",
	}

	mockRepo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil)

	resp, err := authSvc.Login(context.Background(), &model.LoginRequest{
		Email:    user.Email,
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, user.Email, resp.User.Email)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	authSvc, mockRepo := newTestAuthService()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	mockRepo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil)

	resp, err := authSvc.Login(context.Background(), &model.LoginRequest{
		Email:    user.Email,
		Password: "wrongpassword",
	})

	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	authSvc, mockRepo := newTestAuthService()

	mockRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, assert.AnError)

	resp, err := authSvc.Login(context.Background(), &model.LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	})

	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

// --- RefreshToken ---

func TestAuthService_RefreshToken_Success(t *testing.T) {
	authSvc, mockRepo := newTestAuthService()
	userID := uuid.New()
	user := &model.User{ID: userID, Email: "test@example.com", Name: "Test User"}

	// Generate a valid refresh token using the service internals
	svcImpl := authSvc.(*authService)
	refreshToken, err := svcImpl.generateToken(user, "refresh", svcImpl.refreshExpiry)
	assert.NoError(t, err)

	mockRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

	resp, err := authSvc.RefreshToken(context.Background(), refreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, userID, resp.User.ID)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	authSvc, _ := newTestAuthService()

	resp, err := authSvc.RefreshToken(context.Background(), "invalid-token")

	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestAuthService_RefreshToken_UsesAccessToken(t *testing.T) {
	authSvc, mockRepo := newTestAuthService()
	userID := uuid.New()
	user := &model.User{ID: userID, Email: "test@example.com", Name: "Test User"}

	// Generate an ACCESS token, not refresh
	svcImpl := authSvc.(*authService)
	accessToken, err := svcImpl.generateToken(user, "access", svcImpl.jwtExpiry)
	assert.NoError(t, err)

	resp, err := authSvc.RefreshToken(context.Background(), accessToken)

	assert.Nil(t, resp)
	assert.Equal(t, ErrInvalidToken, err)
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestAuthService_RefreshToken_UserNotFound(t *testing.T) {
	authSvc, mockRepo := newTestAuthService()
	userID := uuid.New()
	user := &model.User{ID: userID, Email: "gone@example.com"}

	svcImpl := authSvc.(*authService)
	refreshToken, err := svcImpl.generateToken(user, "refresh", svcImpl.refreshExpiry)
	assert.NoError(t, err)

	mockRepo.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)

	resp, err := authSvc.RefreshToken(context.Background(), refreshToken)

	assert.Nil(t, resp)
	assert.Equal(t, ErrUserNotFound, err)
	mockRepo.AssertExpectations(t)
}
