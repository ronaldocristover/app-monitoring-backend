package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestService_BeforeCreate_WithNilID(t *testing.T) {
	envID := uuid.New()
	serverID := uuid.New()

	svc := Service{
		EnvironmentID: envID,
		ServerID:      serverID,
		Name:          "api-service",
	}

	assert.Equal(t, uuid.Nil, svc.ID)

	err := svc.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, svc.ID)
}

func TestService_BeforeCreate_WithExistingID(t *testing.T) {
	existingID := uuid.New()

	svc := Service{
		ID:   existingID,
		Name: "api-service",
	}

	err := svc.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingID, svc.ID)
}

func TestService_Fields(t *testing.T) {
	envID := uuid.New()
	serverID := uuid.New()

	svc := Service{
		EnvironmentID:  envID,
		ServerID:       serverID,
		Name:           "web-service",
		Type:           "api",
		URL:            "https://example.com",
		Repository:     "https://github.com/example/repo",
		StackLanguage:  "go",
		StackFramework: "gin",
		DBType:         "postgres",
		DBHost:         "db.example.com",
	}

	assert.Equal(t, envID, svc.EnvironmentID)
	assert.Equal(t, serverID, svc.ServerID)
	assert.Equal(t, "web-service", svc.Name)
	assert.Equal(t, "api", svc.Type)
	assert.Equal(t, "https://example.com", svc.URL)
	assert.Equal(t, "https://github.com/example/repo", svc.Repository)
	assert.Equal(t, "go", svc.StackLanguage)
	assert.Equal(t, "gin", svc.StackFramework)
	assert.Equal(t, "postgres", svc.DBType)
	assert.Equal(t, "db.example.com", svc.DBHost)
}

func TestCreateServiceRequest_Fields(t *testing.T) {
	envID := uuid.New()
	serverID := uuid.New()

	req := CreateServiceRequest{
		EnvironmentID:  envID,
		ServerID:       serverID,
		Name:           "svc",
		Type:           "api",
		URL:            "https://example.com",
		Repository:     "https://github.com/example/repo",
		StackLanguage:  "go",
		StackFramework: "gin",
		DBType:         "postgres",
		DBHost:         "db.example.com",
	}

	assert.Equal(t, envID, req.EnvironmentID)
	assert.Equal(t, serverID, req.ServerID)
	assert.Equal(t, "svc", req.Name)
}

func TestUpdateServiceRequest_Fields(t *testing.T) {
	envID := uuid.New()
	serverID := uuid.New()

	req := UpdateServiceRequest{
		EnvironmentID:  &envID,
		ServerID:       &serverID,
		Name:           "updated-svc",
		Type:           "worker",
	}

	assert.Equal(t, &envID, req.EnvironmentID)
	assert.Equal(t, &serverID, req.ServerID)
	assert.Equal(t, "updated-svc", req.Name)
	assert.Equal(t, "worker", req.Type)
}

func TestListServicesRequest_Fields(t *testing.T) {
	req := ListServicesRequest{
		Page:          1,
		PageSize:      20,
		EnvironmentID: uuid.New().String(),
		ServerID:      uuid.New().String(),
		Search:        "api",
	}

	assert.Equal(t, 1, req.Page)
	assert.Equal(t, 20, req.PageSize)
	assert.Equal(t, "api", req.Search)
}
