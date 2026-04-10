package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ronaldocristover/app-monitoring/internal/model"
)

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &model.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}

	err := repo.Create(context.Background(), user)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &model.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}
	require.NoError(t, repo.Create(context.Background(), user))

	found, err := repo.GetByID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)
	assert.Equal(t, user.Name, found.Name)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &model.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}
	require.NoError(t, repo.Create(context.Background(), user))

	found, err := repo.GetByEmail(context.Background(), "test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, user.Email, found.Email)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.GetByEmail(context.Background(), "nonexistent@example.com")
	assert.Error(t, err)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &model.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}
	require.NoError(t, repo.Create(context.Background(), user))

	user.Name = "Updated Name"
	err := repo.Update(context.Background(), user)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &model.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}
	require.NoError(t, repo.Create(context.Background(), user))

	err := repo.Delete(context.Background(), user.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), user.ID)
	assert.Error(t, err)
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	for i := 0; i < 5; i++ {
		user := &model.User{
			Name:         fmt.Sprintf("User %d", i),
			Email:        fmt.Sprintf("user%d@example.com", i),
			PasswordHash: "hashed",
		}
		require.NoError(t, repo.Create(context.Background(), user))
	}

	users, total, err := repo.List(context.Background(), &model.ListUsersRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, users, 5)
}

func TestUserRepository_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	for i := 0; i < 5; i++ {
		user := &model.User{
			Name:         fmt.Sprintf("User %d", i),
			Email:        fmt.Sprintf("user%d@example.com", i),
			PasswordHash: "hashed",
		}
		require.NoError(t, repo.Create(context.Background(), user))
	}

	users, total, err := repo.List(context.Background(), &model.ListUsersRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, users, 2)
}

func TestUserRepository_List_DefaultPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	users, _, err := repo.List(context.Background(), &model.ListUsersRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, users)
}

func TestUserRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
}

func TestNewUserRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	assert.NotNil(t, repo)
}
