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

func TestBackupCreate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	serviceID := uuid.New()
	backup := &model.Backup{
		ID:        uuid.New(),
		ServiceID: serviceID,
		Enabled:   true,
		Path:      "/backups/db.sql",
		Schedule:  "0 2 * * *",
	}

	svc.On("Create", mock.Anything, mock.MatchedBy(func(req *model.CreateBackupRequest) bool {
		return req.ServiceID == serviceID && req.Path == "/backups/db.sql"
	})).Return(backup, nil)

	body := model.CreateBackupRequest{
		ServiceID: serviceID,
		Path:      "/backups/db.sql",
		Schedule:  "0 2 * * *",
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/services/"+serviceID.String()+"/backups", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", serviceID.String())

	h.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestBackupCreate_InvalidServiceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/services/invalid/backups", bytes.NewReader([]byte(`{}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", "invalid")

	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBackupCreate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	serviceID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/services/"+serviceID.String()+"/backups", bytes.NewReader([]byte(`invalid json`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", serviceID.String())

	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBackupCreate_ServiceNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	serviceID := uuid.New()
	svc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateBackupRequest")).Return(nil, service.ErrServiceNotFound)

	body := model.CreateBackupRequest{ServiceID: serviceID, Path: "/backups/db.sql"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/services/"+serviceID.String()+"/backups", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", serviceID.String())

	h.Create(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestBackupCreate_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	serviceID := uuid.New()
	svc.On("Create", mock.Anything, mock.AnythingOfType("*model.CreateBackupRequest")).Return(nil, errors.New("db error"))

	body := model.CreateBackupRequest{ServiceID: serviceID, Path: "/backups/db.sql"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/services/"+serviceID.String()+"/backups", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", serviceID.String())

	h.Create(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- Get ---

func TestBackupGet_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	id := uuid.New()
	backup := &model.Backup{
		ID:        id,
		ServiceID: uuid.New(),
		Path:      "/backups/db.sql",
	}
	svc.On("GetByID", mock.Anything, id).Return(backup, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/backups/"+id.String(), nil)
	c.AddParam("id", id.String())

	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestBackupGet_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/backups/invalid", nil)
	c.AddParam("id", "invalid")

	h.Get(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBackupGet_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	id := uuid.New()
	svc.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/backups/"+id.String(), nil)
	c.AddParam("id", id.String())

	h.Get(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

// --- Update ---

func TestBackupUpdate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	id := uuid.New()
	backup := &model.Backup{
		ID:     id,
		Path:   "/backups/updated.sql",
		Status: "active",
	}
	svc.On("Update", mock.Anything, id, mock.AnythingOfType("*model.UpdateBackupRequest")).Return(backup, nil)

	body := model.UpdateBackupRequest{
		Path:   "/backups/updated.sql",
		Status: "active",
	}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/backups/"+id.String(), bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", id.String())

	h.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestBackupUpdate_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/backups/invalid", bytes.NewReader([]byte(`{"path":"/x"}`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", "invalid")

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBackupUpdate_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	id := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/backups/"+id.String(), bytes.NewReader([]byte(`invalid json`)))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", id.String())

	h.Update(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBackupUpdate_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	id := uuid.New()
	svc.On("Update", mock.Anything, id, mock.AnythingOfType("*model.UpdateBackupRequest")).Return(nil, service.ErrBackupNotFound)

	body := model.UpdateBackupRequest{Path: "/backups/updated.sql"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/backups/"+id.String(), bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", id.String())

	h.Update(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestBackupUpdate_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	id := uuid.New()
	svc.On("Update", mock.Anything, id, mock.AnythingOfType("*model.UpdateBackupRequest")).Return(nil, errors.New("db error"))

	body := model.UpdateBackupRequest{Path: "/backups/updated.sql"}
	b, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/backups/"+id.String(), bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.AddParam("id", id.String())

	h.Update(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- Delete ---

func TestBackupDelete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(nil)

	r := gin.New()
	r.DELETE("/backups/:id", h.Delete)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/backups/"+id.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	svc.AssertExpectations(t)
}

func TestBackupDelete_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/backups/invalid", nil)
	c.AddParam("id", "invalid")

	h.Delete(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBackupDelete_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(service.ErrBackupNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/backups/"+id.String(), nil)
	c.AddParam("id", id.String())

	h.Delete(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestBackupDelete_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	id := uuid.New()
	svc.On("Delete", mock.Anything, id).Return(errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/backups/"+id.String(), nil)
	c.AddParam("id", id.String())

	h.Delete(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}

// --- ListByService ---

func TestBackupListByService_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	serviceID := uuid.New()
	backups := []*model.Backup{
		{ID: uuid.New(), ServiceID: serviceID, Path: "/backups/db1.sql"},
		{ID: uuid.New(), ServiceID: serviceID, Path: "/backups/db2.sql"},
	}

	svc.On("ListByService", mock.Anything, serviceID, mock.AnythingOfType("*model.ListBackupsRequest")).
		Return(backups, int64(2), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/backups?page=1&page_size=20", nil)
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

func TestBackupListByService_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/invalid/backups", nil)
	c.AddParam("id", "invalid")

	h.ListByService(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBackupListByService_ServiceNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	serviceID := uuid.New()
	svc.On("ListByService", mock.Anything, serviceID, mock.AnythingOfType("*model.ListBackupsRequest")).
		Return([]*model.Backup(nil), int64(0), service.ErrServiceNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/backups", nil)
	c.AddParam("id", serviceID.String())

	h.ListByService(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}

func TestBackupListByService_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := new(mocks.BackupService)
	h := NewBackupHandler(svc)

	serviceID := uuid.New()
	svc.On("ListByService", mock.Anything, serviceID, mock.AnythingOfType("*model.ListBackupsRequest")).
		Return([]*model.Backup(nil), int64(0), errors.New("db error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/services/"+serviceID.String()+"/backups", nil)
	c.AddParam("id", serviceID.String())

	h.ListByService(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	svc.AssertExpectations(t)
}
