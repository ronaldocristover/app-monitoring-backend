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

func TestServerRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)
	ctx := context.Background()

	server := &model.Server{
		Name:     "web-01",
		IP:       "192.168.1.1",
		Provider: "aws",
	}

	err := repo.Create(ctx, server)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, server.ID)
}

func TestServerRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)
	ctx := context.Background()

	server := &model.Server{
		Name:     "web-02",
		IP:       "192.168.1.2",
		Provider: "gcp",
	}
	require.NoError(t, repo.Create(ctx, server))

	found, err := repo.GetByID(ctx, server.ID)
	assert.NoError(t, err)
	assert.Equal(t, server.ID, found.ID)
	assert.Equal(t, "web-02", found.Name)
	assert.Equal(t, "192.168.1.2", found.IP)
	assert.Equal(t, "gcp", found.Provider)

	// Non-existent ID returns error
	_, err = repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
}

func TestServerRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)
	ctx := context.Background()

	server := &model.Server{
		Name:     "web-03",
		IP:       "192.168.1.3",
		Provider: "aws",
	}
	require.NoError(t, repo.Create(ctx, server))

	server.Name = "web-03-updated"
	server.IP = "10.0.0.99"
	err := repo.Update(ctx, server)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, server.ID)
	assert.NoError(t, err)
	assert.Equal(t, "web-03-updated", found.Name)
	assert.Equal(t, "10.0.0.99", found.IP)
}

func TestServerRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)
	ctx := context.Background()

	server := &model.Server{
		Name: "web-04",
		IP:   "192.168.1.4",
	}
	require.NoError(t, repo.Create(ctx, server))

	err := repo.Delete(ctx, server.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, server.ID)
	assert.Error(t, err)
}

func TestServerRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)
	ctx := context.Background()

	_, beforeCount, err := repo.List(ctx, &model.ListServersRequest{Page: 1, PageSize: 100})
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		require.NoError(t, repo.Create(ctx, &model.Server{
			Name:     fmt.Sprintf("srv-%d", i),
			IP:       fmt.Sprintf("10.0.0.%d", i),
			Provider: "aws",
		}))
	}

	servers, total, err := repo.List(ctx, &model.ListServersRequest{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, beforeCount+3, total)
	assert.Equal(t, beforeCount+3, int64(len(servers)))
}

func TestServerRepository_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		require.NoError(t, repo.Create(ctx, &model.Server{
			Name: fmt.Sprintf("page-srv-%d", i),
			IP:   fmt.Sprintf("10.0.1.%d", i),
		}))
	}

	servers, total, err := repo.List(ctx, &model.ListServersRequest{Page: 1, PageSize: 2})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, servers, 2)
}

func TestServerRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE requires PostgreSQL")
}
