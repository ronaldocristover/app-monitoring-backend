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
	ctx := context.Background()

	user := &model.User{
		Name:         "Alice",
		Email:        "alice-create@test.com",
		PasswordHash: "hashedpassword",
	}

	err := repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Name:         "Bob",
		Email:        "bob-getbyid@test.com",
		PasswordHash: "hashedpassword",
	}
	require.NoError(t, repo.Create(ctx, user))

	found, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "Bob", found.Name)
	assert.Equal(t, "bob-getbyid@test.com", found.Email)

	// Non-existent ID returns error
	_, err = repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Name:         "Carol",
		Email:        "carol-getbyemail@test.com",
		PasswordHash: "hashedpassword",
	}
	require.NoError(t, repo.Create(ctx, user))

	found, err := repo.GetByEmail(ctx, "carol-getbyemail@test.com")
	assert.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "Carol", found.Name)

	// Non-existent email returns error
	_, err = repo.GetByEmail(ctx, "nonexistent@test.com")
	assert.Error(t, err)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Name:         "Dave",
		Email:        "dave-update@test.com",
		PasswordHash: "hashedpassword",
	}
	require.NoError(t, repo.Create(ctx, user))

	user.Name = "Dave Updated"
	err := repo.Update(ctx, user)
	assert.NoError(t, err)

	updated, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Dave Updated", updated.Name)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Name:         "Eve",
		Email:        "eve-delete@test.com",
		PasswordHash: "hashedpassword",
	}
	require.NoError(t, repo.Create(ctx, user))

	err := repo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, user.ID)
	assert.Error(t, err)
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	_, beforeCount, err := repo.List(ctx, &model.ListUsersRequest{Page: 1, PageSize: 100})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		require.NoError(t, repo.Create(ctx, &model.User{
			Name:         "List User",
			Email:        fmt.Sprintf("list-user-%d@test.com", i),
			PasswordHash: "hashedpassword",
		}))
	}

	users, total, err := repo.List(ctx, &model.ListUsersRequest{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, beforeCount+3, total)
	assert.Equal(t, beforeCount+3, int64(len(users)))
}

func TestUserRepository_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		require.NoError(t, repo.Create(ctx, &model.User{
			Name:         "Page User",
			Email:        fmt.Sprintf("page-user-%d@test.com", i),
			PasswordHash: "hashedpassword",
		}))
	}

	users, total, err := repo.List(ctx, &model.ListUsersRequest{Page: 1, PageSize: 2})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, users, 2)
}

func TestUserRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE requires PostgreSQL")
}
