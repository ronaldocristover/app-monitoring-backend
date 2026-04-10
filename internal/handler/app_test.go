package handler

import (
	"bytes"
	"encoding/json"
	"errors"
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

// --- AppCreate ---

func TestAppCreate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	app := &model.App{ID: uuid.New(), AppName: "myapp"}
	mockAppSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateAppRequest")).Return(app, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"app_name":"myapp"}`
	req, _ := http.NewRequest(http.MethodPost, "/apps", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.SuccessResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	mockAppSvc.AssertExpectations(t)
}

func TestAppCreate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, "/apps", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAppCreate_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	mockAppSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateAppRequest")).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"app_name":"myapp"}`
	req, _ := http.NewRequest(http.MethodPost, "/apps", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Create(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAppSvc.AssertExpectations(t)
}

// --- AppList ---

func TestAppList_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	apps := []*model.App{
		{ID: uuid.New(), AppName: "app1"},
		{ID: uuid.New(), AppName: "app2"},
	}
	mockAppSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListAppsRequest) bool {
		return req.Page == 1 && req.PageSize == 10
	})).Return(apps, int64(2), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/apps?page=1&page_size=10", nil)
	c.Request = req

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.PaginatedResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Meta)
	assert.Equal(t, int64(2), resp.Meta.TotalItems)

	mockAppSvc.AssertExpectations(t)
}

func TestAppList_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	var emptyApps []*model.App
	mockAppSvc.On("List", mock.Anything, mock.MatchedBy(func(req *model.ListAppsRequest) bool {
		return req.Page == 1 && req.PageSize == 20
	})).Return(emptyApps, int64(0), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/apps", nil)
	c.Request = req

	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.PaginatedResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotNil(t, resp.Meta)
	assert.Equal(t, 1, resp.Meta.Page)
	assert.Equal(t, 20, resp.Meta.PageSize)

	mockAppSvc.AssertExpectations(t)
}

func TestAppList_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	mockAppSvc.On("List", mock.Anything, mock.AnythingOfType("*model.ListAppsRequest")).Return([]*model.App{}, int64(0), errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/apps", nil)
	c.Request = req

	h.List(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAppSvc.AssertExpectations(t)
}

// --- AppGet ---

func TestAppGet_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	app := &model.App{ID: appID, AppName: "myapp"}
	mockAppSvc.On("GetByID", mock.Anything, appID).Return(app, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/apps/%s", appID), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.SuccessResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	mockAppSvc.AssertExpectations(t)
}

func TestAppGet_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/apps/not-a-uuid", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Get(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAppGet_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	mockAppSvc.On("GetByID", mock.Anything, appID).Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/apps/%s", appID), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.Get(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockAppSvc.AssertExpectations(t)
}

// --- AppGetDetail ---

func TestAppGetDetail_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	app := &model.App{ID: appID, AppName: "myapp", Environments: []model.Environment{}}
	mockAppSvc.On("GetByIDFull", mock.Anything, appID).Return(app, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/apps/%s/detail", appID), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.GetDetail(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.SuccessResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	mockAppSvc.AssertExpectations(t)
}

func TestAppGetDetail_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/apps/not-a-uuid/detail", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.GetDetail(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAppGetDetail_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	mockAppSvc.On("GetByIDFull", mock.Anything, appID).Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/apps/%s/detail", appID), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.GetDetail(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockAppSvc.AssertExpectations(t)
}

// --- AppUpdate ---

func TestAppUpdate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	updated := &model.App{ID: appID, AppName: "updated"}
	mockAppSvc.On("Update", mock.Anything, appID, mock.AnythingOfType("*model.UpdateAppRequest")).Return(updated, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"app_name":"updated"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/apps/%s", appID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.SuccessResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	mockAppSvc.AssertExpectations(t)
}

func TestAppUpdate_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"app_name":"updated"}`
	req, _ := http.NewRequest(http.MethodPut, "/apps/not-a-uuid", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAppUpdate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/apps/%s", appID), bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAppUpdate_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	mockAppSvc.On("Update", mock.Anything, appID, mock.AnythingOfType("*model.UpdateAppRequest")).Return(nil, service.ErrAppNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"app_name":"updated"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/apps/%s", appID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockAppSvc.AssertExpectations(t)
}

func TestAppUpdate_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	mockAppSvc.On("Update", mock.Anything, appID, mock.AnythingOfType("*model.UpdateAppRequest")).Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"app_name":"updated"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/apps/%s", appID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.Update(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAppSvc.AssertExpectations(t)
}

// --- AppDelete ---

func TestAppDelete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	mockAppSvc.On("Delete", mock.Anything, appID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/apps/%s", appID), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	mockAppSvc.AssertExpectations(t)
}

func TestAppDelete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodDelete, "/apps/not-a-uuid", nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAppDelete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	mockAppSvc.On("Delete", mock.Anything, appID).Return(service.ErrAppNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/apps/%s", appID), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockAppSvc.AssertExpectations(t)
}

func TestAppDelete_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	mockAppSvc.On("Delete", mock.Anything, appID).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/apps/%s", appID), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.Delete(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAppSvc.AssertExpectations(t)
}

// --- AppCreateFull ---

func TestAppCreateFull_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	app := &model.App{ID: uuid.New(), AppName: "fullapp"}
	mockAppSvc.On("CreateFull", mock.Anything, mock.AnythingOfType("*model.CreateFullAppRequest")).Return(app, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"app_name":"fullapp","environments":[{"name":"prod","services":[{"server_id":"` + uuid.New().String() + `","name":"svc1"}]}]}`
	req, _ := http.NewRequest(http.MethodPost, "/apps/full", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.CreateFull(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.SuccessResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	mockAppSvc.AssertExpectations(t)
}

func TestAppCreateFull_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPost, "/apps/full", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.CreateFull(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAppCreateFull_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	mockAppSvc.On("CreateFull", mock.Anything, mock.AnythingOfType("*model.CreateFullAppRequest")).Return(nil, errors.New("tx error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"app_name":"fullapp"}`
	req, _ := http.NewRequest(http.MethodPost, "/apps/full", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.CreateFull(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAppSvc.AssertExpectations(t)
}

// --- AppUpdateFull ---

func TestAppUpdateFull_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	updated := &model.App{ID: appID, AppName: "updated-full"}
	mockAppSvc.On("UpdateFull", mock.Anything, appID, mock.AnythingOfType("*model.UpdateFullAppRequest")).Return(updated, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"app_name":"updated-full"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/apps/%s/full", appID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.UpdateFull(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.SuccessResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp.Success)

	mockAppSvc.AssertExpectations(t)
}

func TestAppUpdateFull_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"app_name":"x"}`
	req, _ := http.NewRequest(http.MethodPut, "/apps/not-a-uuid/full", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	h.UpdateFull(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAppUpdateFull_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/apps/%s/full", appID), bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.UpdateFull(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAppUpdateFull_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	mockAppSvc.On("UpdateFull", mock.Anything, appID, mock.AnythingOfType("*model.UpdateFullAppRequest")).Return(nil, service.ErrAppNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"app_name":"x"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/apps/%s/full", appID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.UpdateFull(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockAppSvc.AssertExpectations(t)
}

func TestAppUpdateFull_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAppSvc := new(mocks.AppService)
	h := NewAppHandler(mockAppSvc)

	appID := uuid.New()
	mockAppSvc.On("UpdateFull", mock.Anything, appID, mock.AnythingOfType("*model.UpdateFullAppRequest")).Return(nil, errors.New("tx error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{"app_name":"x"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/apps/%s/full", appID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: appID.String()}}

	h.UpdateFull(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAppSvc.AssertExpectations(t)
}
