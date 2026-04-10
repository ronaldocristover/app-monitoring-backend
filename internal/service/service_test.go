package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/ronaldocristover/app-monitoring/internal/model"
	"github.com/ronaldocristover/app-monitoring/pkg/apierror"
	"go.uber.org/zap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

func TestServiceCreate_Success(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	envID := uuid.New()
	serverID := uuid.New()
	req := &model.CreateServiceRequest{
		EnvironmentID: envID,
		ServerID:      serverID,
		Name:          "test-service",
		URL:           "http://example.com",
	}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(s *model.Service) bool {
		return s.Name == "test-service" && s.EnvironmentID == envID
	})).Return(nil)

	result, err := svc.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-service", result.Name)
	assert.Equal(t, envID, result.EnvironmentID)
	repo.AssertExpectations(t)
}

func TestServiceCreate_Error(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	req := &model.CreateServiceRequest{
		EnvironmentID: uuid.New(),
		ServerID:      uuid.New(),
		Name:          "test-service",
	}

	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	result, err := svc.Create(context.Background(), req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	repo.AssertExpectations(t)
}

func TestServiceGetByID_Success(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	expected := &model.Service{ID: id, Name: "test-service"}

	repo.On("GetByID", mock.Anything, id).Return(expected, nil)

	result, err := svc.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestServiceGetByID_NotFound(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	result, err := svc.GetByID(context.Background(), id)

	assert.Nil(t, result)
	assert.Equal(t, ErrServiceNotFound, err)
	repo.AssertExpectations(t)
}

func TestServiceGetByIDFull_Success(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	expected := &model.Service{ID: id, Name: "test-service"}

	repo.On("GetByIDFull", mock.Anything, id).Return(expected, nil)

	result, err := svc.GetByIDFull(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestServiceGetByIDFull_NotFound(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	repo.On("GetByIDFull", mock.Anything, id).Return(nil, errors.New("not found"))

	result, err := svc.GetByIDFull(context.Background(), id)

	assert.Nil(t, result)
	assert.Equal(t, ErrServiceNotFound, err)
	repo.AssertExpectations(t)
}

func TestServiceUpdate_Success(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Service{ID: id, Name: "old-name", URL: "http://old.com"}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(s *model.Service) bool {
		return s.Name == "new-name" && s.URL == "http://new.com"
	})).Return(nil)

	req := &model.UpdateServiceRequest{
		Name: "new-name",
		URL:  "http://new.com",
	}
	result, err := svc.Update(context.Background(), id, req)

	assert.NoError(t, err)
	assert.Equal(t, "new-name", result.Name)
	assert.Equal(t, "http://new.com", result.URL)
	repo.AssertExpectations(t)
}

func TestServiceUpdate_NotFound(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	req := &model.UpdateServiceRequest{Name: "new-name"}
	result, err := svc.Update(context.Background(), id, req)

	assert.Nil(t, result)
	assert.Equal(t, ErrServiceNotFound, err)
	repo.AssertExpectations(t)
}

func TestServiceUpdate_RepoError(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Service{ID: id, Name: "old-name"}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))

	req := &model.UpdateServiceRequest{Name: "new-name"}
	result, err := svc.Update(context.Background(), id, req)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	repo.AssertExpectations(t)
}

func TestServiceDelete_Success(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Service{ID: id}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Delete", mock.Anything, id).Return(nil)

	err := svc.Delete(context.Background(), id)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceDelete_NotFound(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	err := svc.Delete(context.Background(), id)

	assert.Equal(t, ErrServiceNotFound, err)
	repo.AssertExpectations(t)
}

func TestServiceDelete_RepoError(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Service{ID: id}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	repo.On("Delete", mock.Anything, id).Return(errors.New("db error"))

	err := svc.Delete(context.Background(), id)

	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	repo.AssertExpectations(t)
}

func TestServiceList_Success(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	services := []*model.Service{
		{ID: uuid.New(), Name: "svc1"},
		{ID: uuid.New(), Name: "svc2"},
	}
	req := &model.ListServicesRequest{Page: 1, PageSize: 10}

	repo.On("List", mock.Anything, req).Return(services, int64(2), nil)

	result, total, err := svc.List(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	repo.AssertExpectations(t)
}

func TestServiceManualPing_Success(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	id := uuid.New()
	existing := &model.Service{ID: id, URL: ts.URL}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	logRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *model.MonitoringLog) bool {
		return log.ServiceID == id && log.Status == "up" && log.StatusCode == 200
	})).Return(nil)

	result, err := svc.ManualPing(context.Background(), id)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "up", result.Status)
	assert.Equal(t, 200, result.StatusCode)
	repo.AssertExpectations(t)
	logRepo.AssertExpectations(t)
}

func TestServiceManualPing_ServiceNotFound(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	repo.On("GetByID", mock.Anything, id).Return(nil, errors.New("not found"))

	result, err := svc.ManualPing(context.Background(), id)

	assert.Nil(t, result)
	assert.Equal(t, ErrServiceNotFound, err)
	repo.AssertExpectations(t)
}

func TestServiceManualPing_NoURL(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Service{ID: id, URL: ""}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)

	result, err := svc.ManualPing(context.Background(), id)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	repo.AssertExpectations(t)
}

func TestServiceManualPing_DownStatus(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	id := uuid.New()
	existing := &model.Service{ID: id, URL: ts.URL}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	logRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *model.MonitoringLog) bool {
		return log.Status == "down" && log.StatusCode == 500
	})).Return(nil)

	result, err := svc.ManualPing(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, "down", result.Status)
	assert.Equal(t, 500, result.StatusCode)
	repo.AssertExpectations(t)
	logRepo.AssertExpectations(t)
}

func TestServiceManualPing_RequestError(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	id := uuid.New()
	existing := &model.Service{ID: id, URL: "http://localhost:1"}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	logRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *model.MonitoringLog) bool {
		return log.Status == "down" && log.ErrorMessage != ""
	})).Return(nil)

	result, err := svc.ManualPing(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, "down", result.Status)
	assert.NotEmpty(t, result.ErrorMessage)
	repo.AssertExpectations(t)
	logRepo.AssertExpectations(t)
}

func TestServiceManualPing_SaveLogError(t *testing.T) {
	repo := new(MockServiceRepository)
	logRepo := new(MockMonitoringLogRepository)
	svc := NewServiceService(repo, logRepo, newTestLogger())

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	id := uuid.New()
	existing := &model.Service{ID: id, URL: ts.URL}

	repo.On("GetByID", mock.Anything, id).Return(existing, nil)
	logRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	result, err := svc.ManualPing(context.Background(), id)

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.IsType(t, &apierror.Error{}, err)
	repo.AssertExpectations(t)
	logRepo.AssertExpectations(t)
}
