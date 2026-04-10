package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestApp_BeforeCreate_WithNilID(t *testing.T) {
	app := App{
		AppName:     "my-app",
		Description: "A test app",
	}

	assert.Equal(t, uuid.Nil, app.ID)

	err := app.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, app.ID)
}

func TestApp_BeforeCreate_WithExistingID(t *testing.T) {
	existingID := uuid.New()
	app := App{
		ID:      existingID,
		AppName: "my-app",
	}

	err := app.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingID, app.ID)
}

func TestCreateAppRequest_Fields(t *testing.T) {
	req := CreateAppRequest{
		AppName:     "test-app",
		Description: "desc",
		Tags:        "api,web",
	}

	assert.Equal(t, "test-app", req.AppName)
	assert.Equal(t, "desc", req.Description)
	assert.Equal(t, "api,web", req.Tags)
}

func TestUpdateAppRequest_Fields(t *testing.T) {
	req := UpdateAppRequest{
		AppName:     "updated-app",
		Description: "updated",
		Tags:        "api",
	}

	assert.Equal(t, "updated-app", req.AppName)
	assert.Equal(t, "updated", req.Description)
	assert.Equal(t, "api", req.Tags)
}

func TestListAppsRequest_Fields(t *testing.T) {
	req := ListAppsRequest{
		Page:     1,
		PageSize: 20,
		Search:   "myapp",
		Tags:     "api",
	}

	assert.Equal(t, 1, req.Page)
	assert.Equal(t, 20, req.PageSize)
	assert.Equal(t, "myapp", req.Search)
	assert.Equal(t, "api", req.Tags)
}

func TestApp_Fields(t *testing.T) {
	app := App{
		AppName:     "my-app",
		Description: "description",
		Tags:        "tag1,tag2",
	}

	assert.Equal(t, "my-app", app.AppName)
	assert.Equal(t, "description", app.Description)
	assert.Equal(t, "tag1,tag2", app.Tags)
}
