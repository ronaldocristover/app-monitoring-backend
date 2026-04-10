package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ronaldocristover/app-monitoring/internal/handler/mocks"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/internal/service"
	"github.com/ronaldocristover/app-monitoring/pkg/response"
)

// --- Register ---

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	loginResp := &model.LoginResponse{
		Token:        "access-token",
		RefreshToken: "refresh-token",
		User:         model.User{ID: uuid.New(), Name: "John Doe", Email: "john@example.com"},
	}
	mockAuthSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).Return(loginResp, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"email":"john@example.com","password":"password123","name":"John Doe"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Register(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.SuccessResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	mockAuthSvc.AssertExpectations(t)
}

func TestRegister_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Register(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_UserExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	mockAuthSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).Return(nil, service.ErrUserExists)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"email":"john@example.com","password":"password123","name":"John Doe"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Register(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockAuthSvc.AssertExpectations(t)
}

func TestRegister_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	mockAuthSvc.On("Register", mock.Anything, mock.AnythingOfType("*model.RegisterRequest")).Return(nil, errors.New("unexpected error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"email":"john@example.com","password":"password123","name":"John Doe"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Register(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAuthSvc.AssertExpectations(t)
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	loginResp := &model.LoginResponse{
		Token:        "access-token",
		RefreshToken: "refresh-token",
		User:         model.User{ID: uuid.New(), Name: "John Doe", Email: "john@example.com"},
	}
	mockAuthSvc.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).Return(loginResp, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"email":"john@example.com","password":"password123"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.SuccessResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	mockAuthSvc.AssertExpectations(t)
}

func TestLogin_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	mockAuthSvc.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).Return(nil, service.ErrInvalidCredentials)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"email":"john@example.com","password":"wrongpassword"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthSvc.AssertExpectations(t)
}

func TestLogin_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	mockAuthSvc.On("Login", mock.Anything, mock.AnythingOfType("*model.LoginRequest")).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"email":"john@example.com","password":"password123"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Login(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAuthSvc.AssertExpectations(t)
}

// --- Me ---

func TestMe_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	userID := uuid.New()
	user := &model.User{ID: userID, Name: "John Doe", Email: "john@example.com"}
	mockUserSvc.On("GetByID", mock.Anything, userID).Return(user, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/auth/me", nil)
	c.Request = req
	c.Set("userID", userID)

	h.Me(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.SuccessResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	mockUserSvc.AssertExpectations(t)
}

func TestMe_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/auth/me", nil)
	c.Request = req

	h.Me(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMe_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/auth/me", nil)
	c.Request = req
	c.Set("userID", "not-a-uuid")

	h.Me(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMe_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	userID := uuid.New()
	mockUserSvc.On("GetByID", mock.Anything, userID).Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/auth/me", nil)
	c.Request = req
	c.Set("userID", userID)

	h.Me(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUserSvc.AssertExpectations(t)
}

// --- RefreshToken ---

func TestRefreshToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	loginResp := &model.LoginResponse{
		Token:        "new-access-token",
		RefreshToken: "new-refresh-token",
		User:         model.User{ID: uuid.New(), Name: "John Doe", Email: "john@example.com"},
	}
	mockAuthSvc.On("RefreshToken", mock.Anything, "valid-refresh-token").Return(loginResp, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"refresh_token":"valid-refresh-token"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.RefreshToken(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.SuccessResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	mockAuthSvc.AssertExpectations(t)
}

func TestRefreshToken_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.RefreshToken(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	mockAuthSvc.On("RefreshToken", mock.Anything, "expired-token").Return(nil, service.ErrInvalidToken)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"refresh_token":"expired-token"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.RefreshToken(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthSvc.AssertExpectations(t)
}

func TestRefreshToken_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	mockAuthSvc.On("RefreshToken", mock.Anything, "valid-token-deleted-user").Return(nil, service.ErrUserNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"refresh_token":"valid-token-deleted-user"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.RefreshToken(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthSvc.AssertExpectations(t)
}

func TestRefreshToken_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(mocks.AuthService)
	mockUserSvc := new(mocks.UserService)
	h := NewAuthHandler(mockAuthSvc, mockUserSvc)

	mockAuthSvc.On("RefreshToken", mock.Anything, "some-token").Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"refresh_token":"some-token"}`
	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.RefreshToken(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAuthSvc.AssertExpectations(t)
}
