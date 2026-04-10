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

// --- Create ---

func TestDeploymentCreate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	serviceID := uuid.New()
	deployment := &model.Deployment{
		ID:            uuid.New(),
		ServiceID:     serviceID,
		Method:        "docker",
		ContainerName: "my-container",
		Port:          8080,
	}

	svc.On("Create", mock.Anything, mock.MatchedBy(func(req *model.CreateDeploymentRequest) bool {
		return req.ServiceID == serviceID && req.Method == "docker"
	})).Return(deployment, nil)

	body := model.CreateDeploymentRequest{
		ServiceID:     serviceID,
		Method:        "docker",
		ContainerName: "my-container",
		Port:          8080,
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/services/"+serviceID.String()+"/deployments", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", serviceID.String())

	h.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestDeploymentCreate_InvalidServiceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/services/invalid/deployments", bytes.NewReader([]byte(`{}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", "invalid")

	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeploymentCreate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	serviceID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/services/"+serviceID.String()+"/deployments", bytes.NewReader([]byte(`invalid json`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", serviceID.String())

	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeploymentCreate_ServiceNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	serviceID := uuid.New()
	svc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateDeploymentRequest")).Return(nil, service.ErrServiceNotFound)

	body := model.CreateDeploymentRequest{
		ServiceID: serviceID,
		Method:    "docker",
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/services/"+serviceID.String()+"/deployments", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", serviceID.String())

	h.Create(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestDeploymentCreate_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	serviceID := uuid.New()
	svc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateDeploymentRequest")).Return(nil, errors.New("db error"))

	body := model.CreateDeploymentRequest{
		ServiceID: serviceID,
		Method:    "docker",
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/services/"+serviceID.String()+"/deployments", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", serviceID.String())

	h.Create(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- Get ---

func TestDeploymentGet_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	id := uuid.New()
	deployment := &model.Deployment{
		ID:        id,
		ServiceID: uuid.New(),
		Method:    "docker",
	}
	svc.On("GetByID", mock.Anything, id).Return(deployment, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/deployments/"+id.String(), nil)
	c.AddParam("id", id.String())

	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestDeploymentGet_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/deployments/invalid", nil)
	c.AddParam("id", "invalid")

	h.Get(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeploymentGet_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	id := uuid.New()
	svc.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/deployments/"+id.String(), nil)
	c.AddParam("id", id.String())

	h.Get(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

// --- Update ---

func TestDeploymentUpdate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	id := uuid.New()
	deployment := &model.Deployment{
		ID:     id,
		Method: "kubernetes",
	}
	svc.On("Update", mock.Anything, id, mock.AnythingOfType("*model.UpdateDeploymentRequest")).Return(deployment, nil)

	body := model.UpdateDeploymentRequest{
		Method: "kubernetes",
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/deployments/"+id.String(), bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", id.String())

	h.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestDeploymentUpdate_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/deployments/invalid", bytes.NewReader([]byte(`{"method":"docker"}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", "invalid")

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeploymentUpdate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	id := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/deployments/"+id.String(), bytes.NewReader([]byte(`invalid json`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", id.String())

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeploymentUpdate_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	id := uuid.New()
	svc.On("Update", mock.Anything, id, mock.AnythingOfType("*model.UpdateDeploymentRequest")).Return(nil, service.ErrDeploymentNotFound)

	body := model.UpdateDeploymentRequest{
		Method: "kubernetes",
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/deployments/"+id.String(), bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", id.String())

	h.Update(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestDeploymentUpdate_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	id := uuid.New()
	svc.On("Update", mock.Anything, id, mock.AnythingOfType("*model.UpdateDeploymentRequest")).Return(nil, errors.New("db error"))

	body := model.UpdateDeploymentRequest{
		Method: "kubernetes",
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/deployments/"+id.String(), bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", id.String())

	h.Update(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- Delete ---

func TestDeploymentDelete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(nil)

	r := gin.New()
	r.DELETE("/deployments/:id", h.Delete)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/deployments/"+id.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	svc.AssertExpectations(t)
}

func TestDeploymentDelete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/deployments/invalid", nil)
	c.AddParam("id", "invalid")

	h.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeploymentDelete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(service.ErrDeploymentNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/deployments/"+id.String(), nil)
	c.AddParam("id", id.String())

	h.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestDeploymentDelete_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/deployments/"+id.String(), nil)
	c.AddParam("id", id.String())

	h.Delete(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- ListByService ---

func TestDeploymentListByService_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	serviceID := uuid.New()
	deployments := []*model.Deployment{
		{ID: uuid.New(), ServiceID: serviceID, Method: "docker"},
		{ID: uuid.New(), ServiceID: serviceID, Method: "kubernetes"},
	}

	svc.On("ListByService", mock.Anything, serviceID, mock.AnythingOfType("*model.ListDeploymentsRequest")).
		Return(deployments, int64(2), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/deployments?page=1&page_size=20", nil)
	c.AddParam("id", serviceID.String())

	h.ListByService(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Meta)

	svc.AssertExpectations(t)
}

func TestDeploymentListByService_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/invalid/deployments", nil)
	c.AddParam("id", "invalid")

	h.ListByService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeploymentListByService_ServiceNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	serviceID := uuid.New()
	svc.On("ListByService", mock.Anything, serviceID, mock.AnythingOfType("*model.ListDeploymentsRequest")).
		Return([]*model.Deployment(nil), int64(0), service.ErrServiceNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/deployments", nil)
	c.AddParam("id", serviceID.String())

	h.ListByService(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestDeploymentListByService_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.DeploymentService)
	h := NewDeploymentHandler(svc)

	serviceID := uuid.New()
	svc.On("ListByService", mock.Anything, serviceID, mock.AnythingOfType("*model.ListDeploymentsRequest")).
		Return([]*model.Deployment(nil), int64(0), errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/deployments", nil)
	c.AddParam("id", serviceID.String())

	h.ListByService(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}
