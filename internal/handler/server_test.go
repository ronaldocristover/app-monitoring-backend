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

// --- Create ---

func TestServerCreate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	srv := &model.Server{ID: uuid.New(), Name: "web-01", IP: "10.0.0.1", Provider: "aws"}
	svc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateServerRequest")).Return(srv, nil)

	body := model.CreateServerRequest{Name: "web-01", IP: "10.0.0.1", Provider: "aws"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/servers", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestServerCreate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/servers", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServerCreate_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	svc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateServerRequest")).Return(nil, assert.AnError)

	body := model.CreateServerRequest{Name: "web-01", IP: "10.0.0.1"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/servers", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- List ---

func TestServerList_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	servers := []*model.Server{
		{ID: uuid.New(), Name: "web-01", IP: "10.0.0.1"},
		{ID: uuid.New(), Name: "web-02", IP: "10.0.0.2"},
	}
	svc.On("List", mock.Anything, mock.AnythingOfType("*model.ListServersRequest")).Return(servers, int64(2), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/servers", nil)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.PaginatedResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, int64(2), resp.Meta.TotalItems)
	svc.AssertExpectations(t)
}

func TestServerList_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	var emptyServers []*model.Server
	svc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListServersRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(emptyServers, int64(0), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/servers", nil)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestServerList_WithPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	var emptyServers []*model.Server
	svc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListServersRequest) bool {
		return req.Page == 2 && req.PageSize == 5
	})).Return(emptyServers, int64(8), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/servers?page=2&page_size=5", nil)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp response.PaginatedResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, int64(8), resp.Meta.TotalItems)
	svc.AssertExpectations(t)
}

func TestServerList_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	svc.On("List", mock.Anything, mock.AnythingOfType("*model.ListServersRequest")).Return([]*model.Server{}, int64(0), assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/servers", nil)

	h.List(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- Get ---

func TestServerGet_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	serverID := uuid.New()
	srv := &model.Server{ID: serverID, Name: "web-01", IP: "10.0.0.1", Provider: "aws"}
	svc.On("GetByID", mock.Anything, serverID).Return(srv, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/servers/"+serverID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: serverID.String()}}

	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestServerGet_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/servers/not-a-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Get(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServerGet_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	serverID := uuid.New()
	svc.On("GetByID", mock.Anything, serverID).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/servers/"+serverID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: serverID.String()}}

	h.Get(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

// --- Update ---

func TestServerUpdate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	serverID := uuid.New()
	updatedServer := &model.Server{ID: serverID, Name: "web-01-updated", IP: "10.0.0.100", Provider: "gcp"}
	svc.On("Update", mock.Anything, serverID, mock.AnythingOfType("*model.UpdateServerRequest")).Return(updatedServer, nil)

	body := model.UpdateServerRequest{Name: "web-01-updated", IP: "10.0.0.100", Provider: "gcp"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/servers/%s", serverID), bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: serverID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestServerUpdate_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	body := model.UpdateServerRequest{Name: "updated"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPut, "/servers/not-a-uuid", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServerUpdate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	serverID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/servers/%s", serverID), bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: serverID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServerUpdate_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	serverID := uuid.New()
	svc.On("Update", mock.Anything, serverID, mock.AnythingOfType("*model.UpdateServerRequest")).Return(nil, service.ErrServerNotFound)

	body := model.UpdateServerRequest{Name: "updated"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/servers/%s", serverID), bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: serverID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestServerUpdate_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	serverID := uuid.New()
	svc.On("Update", mock.Anything, serverID, mock.AnythingOfType("*model.UpdateServerRequest")).Return(nil, assert.AnError)

	body := model.UpdateServerRequest{Name: "updated"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/servers/%s", serverID), bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: serverID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- Delete ---

func TestServerDelete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	serverID := uuid.New()
	svc.On("Delete", mock.Anything, serverID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/servers/%s", serverID), nil)
	c.Params = gin.Params{{Key: "id", Value: serverID.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	svc.AssertExpectations(t)
}

func TestServerDelete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodDelete, "/servers/not-a-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServerDelete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	serverID := uuid.New()
	svc.On("Delete", mock.Anything, serverID).Return(service.ErrServerNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/servers/%s", serverID), nil)
	c.Params = gin.Params{{Key: "id", Value: serverID.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestServerDelete_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.ServerService)
	h := NewServerHandler(svc)

	serverID := uuid.New()
	svc.On("Delete", mock.Anything, serverID).Return(assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/servers/%s", serverID), nil)
	c.Params = gin.Params{{Key: "id", Value: serverID.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}
