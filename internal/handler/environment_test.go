package handler

import (
	"bytes"
	"encoding/json"
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
)

// --- Create ---

func TestEnvCreate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	appID := uuid.New()
	envID := uuid.New()
	env := &model.Environment{ID: envID, AppID: appID, Name: "prod"}
	svc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateEnvironmentRequest")).Return(env, nil)

	body := model.CreateEnvironmentRequest{AppID: appID, Name: "prod"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/environments", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestEnvCreate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/environments", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEnvCreate_InvalidAppID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	svc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateEnvironmentRequest")).Return(nil, service.ErrInvalidAppID)

	body := model.CreateEnvironmentRequest{AppID: uuid.New(), Name: "prod"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/environments", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	svc.AssertExpectations(t)
}

func TestEnvCreate_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	svc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateEnvironmentRequest")).Return(nil, assert.AnError)

	body := model.CreateEnvironmentRequest{AppID: uuid.New(), Name: "prod"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/environments", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- Get ---

func TestEnvGet_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	id := uuid.New()
	appID := uuid.New()
	env := &model.Environment{ID: id, AppID: appID, Name: "prod"}
	svc.On("GetByID", mock.Anything, id).Return(env, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/environments/"+id.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestEnvGet_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/environments/not-a-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Get(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEnvGet_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	id := uuid.New()
	svc.On("GetByID", mock.Anything, id).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/environments/"+id.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Get(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

// --- Update ---

func TestEnvUpdate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	id := uuid.New()
	appID := uuid.New()
	env := &model.Environment{ID: id, AppID: appID, Name: "staging"}
	svc.On("Update", mock.Anything, id, mock.AnythingOfType("*model.UpdateEnvironmentRequest")).Return(env, nil)

	body := model.UpdateEnvironmentRequest{Name: "staging"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/environments/"+id.String(), bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestEnvUpdate_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	body := model.UpdateEnvironmentRequest{Name: "staging"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/environments/not-a-uuid", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEnvUpdate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	id := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/environments/"+id.String(), bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEnvUpdate_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	id := uuid.New()
	svc.On("Update", mock.Anything, id, mock.AnythingOfType("*model.UpdateEnvironmentRequest")).Return(nil, service.ErrEnvironmentNotFound)

	body := model.UpdateEnvironmentRequest{Name: "staging"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/environments/"+id.String(), bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestEnvUpdate_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	id := uuid.New()
	svc.On("Update", mock.Anything, id, mock.AnythingOfType("*model.UpdateEnvironmentRequest")).Return(nil, assert.AnError)

	body := model.UpdateEnvironmentRequest{Name: "staging"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/environments/"+id.String(), bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- Delete ---

func TestEnvDelete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/environments/"+id.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	svc.AssertExpectations(t)
}

func TestEnvDelete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/environments/not-a-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEnvDelete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(service.ErrEnvironmentNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/environments/"+id.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestEnvDelete_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/environments/"+id.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- List ---

func TestEnvList_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	envs := []*model.Environment{{ID: uuid.New(), AppID: uuid.New(), Name: "prod"}}
	svc.On("List", mock.Anything, mock.AnythingOfType("*model.ListEnvironmentsRequest")).Return(envs, int64(1), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/environments?page=1&page_size=20", nil)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["meta"])
	svc.AssertExpectations(t)
}

func TestEnvList_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	envs := []*model.Environment{{ID: uuid.New(), AppID: uuid.New(), Name: "prod"}}
	svc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListEnvironmentsRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(envs, int64(1), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/environments", nil)

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestEnvList_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := new(mocks.EnvironmentService)
	h := NewEnvironmentHandler(svc)

	svc.On("List", mock.Anything, mock.AnythingOfType("*model.ListEnvironmentsRequest")).Return([]*model.Environment(nil), int64(0), assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/environments?page=1&page_size=20", nil)

	h.List(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}
