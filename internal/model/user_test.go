package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_BeforeCreate_WithNilID(t *testing.T) {
	user := User{
		Email:        "test@example.com",
		PasswordHash: "hashed",
		Name:         "Test User",
	}

	assert.Equal(t, uuid.Nil, user.ID)

	err := user.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
}

func TestUser_BeforeCreate_WithExistingID(t *testing.T) {
	existingID := uuid.New()
	user := User{
		ID:           existingID,
		Email:        "test@example.com",
		PasswordHash: "hashed",
		Name:         "Test User",
	}

	err := user.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingID, user.ID)
}

func TestRegisterRequest_Valid(t *testing.T) {
	req := RegisterRequest{
		Email:    "user@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	assert.Equal(t, "user@example.com", req.Email)
	assert.Equal(t, "password123", req.Password)
	assert.Equal(t, "Test User", req.Name)
}

func TestLoginRequest_Fields(t *testing.T) {
	req := LoginRequest{
		Email:    "user@example.com",
		Password: "password123",
	}

	assert.Equal(t, "user@example.com", req.Email)
	assert.Equal(t, "password123", req.Password)
}

func TestRefreshTokenRequest_Field(t *testing.T) {
	req := RefreshTokenRequest{RefreshToken: "token123"}

	assert.Equal(t, "token123", req.RefreshToken)
}

func TestUpdateUserRequest_Fields(t *testing.T) {
	req := UpdateUserRequest{
		Name:  "New Name",
		Email: "new@example.com",
	}

	assert.Equal(t, "New Name", req.Name)
	assert.Equal(t, "new@example.com", req.Email)
}

func TestListUsersRequest_Fields(t *testing.T) {
	req := ListUsersRequest{
		Page:     2,
		PageSize: 10,
		Search:   "john",
	}

	assert.Equal(t, 2, req.Page)
	assert.Equal(t, 10, req.PageSize)
	assert.Equal(t, "john", req.Search)
}

func TestLoginResponse_Fields(t *testing.T) {
	user := User{
		Email: "test@example.com",
		Name:  "Test",
	}
	resp := LoginResponse{
		Token:        "jwt-token",
		RefreshToken: "refresh-token",
		User:         user,
	}

	assert.Equal(t, "jwt-token", resp.Token)
	assert.Equal(t, "refresh-token", resp.RefreshToken)
	assert.Equal(t, "test@example.com", resp.User.Email)
}
