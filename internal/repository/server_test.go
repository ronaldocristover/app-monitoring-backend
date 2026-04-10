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

func createTestServer(t *testing.T, repo ServerRepository, suffix string) *model.Server {
	t.Helper()
	server := &model.Server{
		Name:     fmt.Sprintf("server_%s", suffix),
		IP:       fmt.Sprintf("10.0.0.%s", suffix),
		Provider: "aws",
	}
	require.NoError(t, repo.Create(context.Background(), server))
	return server
}

func TestServerRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)

	server := &model.Server{
		Name:     "web-01",
		IP:       "192.168.1.1",
		Provider: "aws",
	}

	err := repo.Create(context.Background(), server)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, server.ID)
}

func TestServerRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)

	server := createTestServer(t, repo, "01")

	found, err := repo.GetByID(context.Background(), server.ID)
	assert.NoError(t, err)
	assert.Equal(t, server.ID, found.ID)
	assert.Equal(t, server.Name, found.Name)
	assert.Equal(t, server.IP, found.IP)
	assert.Equal(t, server.Provider, found.Provider)
}

func TestServerRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestServerRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)

	server := createTestServer(t, repo, "upd")
	server.Name = "updated-server"
	server.IP = "10.0.0.99"

	err := repo.Update(context.Background(), server)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), server.ID)
	assert.NoError(t, err)
	assert.Equal(t, "updated-server", found.Name)
	assert.Equal(t, "10.0.0.99", found.IP)
}

func TestServerRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)

	server := createTestServer(t, repo, "del")

	err := repo.Delete(context.Background(), server.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), server.ID)
	assert.Error(t, err)
}

func TestServerRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)

	for i := 0; i < 5; i++ {
		server := &model.Server{
			Name:     fmt.Sprintf("srv_%d", i),
			IP:       fmt.Sprintf("10.0.1.%d", i),
			Provider: "aws",
		}
		require.NoError(t, repo.Create(context.Background(), server))
	}

	servers, total, err := repo.List(context.Background(), &model.ListServersRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, servers, 5)
}

func TestServerRepository_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)

	for i := 0; i < 5; i++ {
		server := &model.Server{
			Name: fmt.Sprintf("srv_%d", i),
			IP:   fmt.Sprintf("10.0.2.%d", i),
		}
		require.NoError(t, repo.Create(context.Background(), server))
	}

	servers, total, err := repo.List(context.Background(), &model.ListServersRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, servers, 2)
}

func TestServerRepository_List_WithSearch(t *testing.T) {
	t.Skip("ILIKE is PostgreSQL-specific; this test requires PostgreSQL")
}

func TestNewServerRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepository(db)
	assert.NotNil(t, repo)
}
