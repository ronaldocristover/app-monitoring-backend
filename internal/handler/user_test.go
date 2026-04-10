package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// --- List ---

func TestUserList_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	users := []*model.User{
		{ID: uuid.New(), Name: "John", Email: "john@example.com"},
		{ID: uuid.New(), Name: "Jane", Email: "jane@example.com"},
	}
	svc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListUsersRequest) bool {
		return req.Page == 1 && req.PageSize == 10
	})).Return(users, int64(2), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/users?page=1&page_size=10", nil)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.PaginatedResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, int64(2), resp.Meta.TotalItems)
	svc.AssertExpectations(t)
}

func TestUserList_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	svc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListUsersRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return([]*model.User{}, int64(0), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/users", nil)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.PaginatedResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 1, resp.Meta.Page)
	assert.Equal(t, 20, resp.Meta.PageSize)
	svc.AssertExpectations(t)
}

func TestUserList_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/users?page=-1", nil)

	h.List(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserList_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	svc.On("List", mock.Anything, mock.AnythingOfType("*model.ListUsersRequest")).Return([]*model.User{}, int64(0), assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/users", nil)

	h.List(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- Get ---

func TestUserGet_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	userID := uuid.New()
	user := &model.User{ID: userID, Name: "John Doe", Email: "john@example.com"}
	svc.On("GetByID", mock.Anything, userID).Return(user, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/users/"+userID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}

	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.SuccessResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	svc.AssertExpectations(t)
}

func TestUserGet_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/users/not-a-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Get(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserGet_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	userID := uuid.New()
	svc.On("GetByID", mock.Anything, userID).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s", userID), nil)
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}

	h.Get(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

// --- Update ---

func TestUserUpdate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	userID := uuid.New()
	updatedUser := &model.User{ID: userID, Name: "Updated", Email: "updated@example.com"}
	svc.On("Update", mock.Anything, userID, mock.AnythingOfType("*model.UpdateUserRequest")).Return(updatedUser, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/users/%s", userID), bytes.NewBufferString(
		`{"name":"Updated","email":"updated@example.com"}`,
	))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.SuccessResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	svc.AssertExpectations(t)
}

func TestUserUpdate_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPut, "/users/not-a-uuid", bytes.NewBufferString(`{"name":"Updated"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserUpdate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	userID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/users/%s", userID), bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserUpdate_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	userID := uuid.New()
	svc.On("Update", mock.Anything, userID, mock.AnythingOfType("*model.UpdateUserRequest")).Return(nil, service.ErrUserNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/users/%s", userID), bytes.NewBufferString(`{"name":"Updated"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestUserUpdate_Conflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	userID := uuid.New()
	svc.On("Update", mock.Anything, userID, mock.AnythingOfType("*model.UpdateUserRequest")).Return(nil, service.ErrUserExists)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/users/%s", userID), bytes.NewBufferString(`{"email":"dup@example.com"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusConflict, w.Code)
	svc.AssertExpectations(t)
}

func TestUserUpdate_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	userID := uuid.New()
	svc.On("Update", mock.Anything, userID, mock.AnythingOfType("*model.UpdateUserRequest")).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/users/%s", userID), bytes.NewBufferString(`{"name":"Updated"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- Delete ---

func TestUserDelete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	userID := uuid.New()
	svc.On("Delete", mock.Anything, userID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s", userID), nil)
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	svc.AssertExpectations(t)
}

func TestUserDelete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodDelete, "/users/not-a-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserDelete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	userID := uuid.New()
	svc.On("Delete", mock.Anything, userID).Return(service.ErrUserNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s", userID), nil)
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestUserDelete_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.UserService)
	h := NewUserHandler(svc)

	userID := uuid.New()
	svc.On("Delete", mock.Anything, userID).Return(assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s", userID), nil)
	c.Params = gin.Params{{Key: "id", Value: userID.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}
