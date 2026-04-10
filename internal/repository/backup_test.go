package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ronaldocristover/app-monitoring/internal/model"
	"gorm.io/gorm"
)

func createTestBackup(t *testing.T, db *gorm.DB, repo BackupRepository, suffix string) (*model.Service, *model.Backup) {
	t.Helper()
	app := &model.App{AppName: fmt.Sprintf("BackupApp_%s", suffix)}
	require.NoError(t, db.Create(app).Error)

	env := &model.Environment{AppID: app.ID, Name: "prod"}
	require.NoError(t, db.Create(env).Error)

	server := &model.Server{Name: fmt.Sprintf("srv_%s", suffix), IP: "10.0.0.1"}
	require.NoError(t, db.Create(server).Error)

	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          fmt.Sprintf("backup-svc-%s", suffix),
	}
	require.NoError(t, db.Create(svc).Error)

	now := time.Now()
	backup := &model.Backup{
		ServiceID:      svc.ID,
		Enabled:        true,
		Path:           fmt.Sprintf("/backups/%s", suffix),
		Schedule:       "0 2 * * *",
		LastBackupTime: &now,
		Status:         "success",
	}
	require.NoError(t, repo.Create(context.Background(), backup))
	return svc, backup
}

func TestBackupRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)

	_, backup := createTestBackup(t, db, repo, "create")
	assert.NotEqual(t, uuid.Nil, backup.ID)
}

func TestBackupRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)

	_, backup := createTestBackup(t, db, repo, "get")

	found, err := repo.GetByID(context.Background(), backup.ID)
	assert.NoError(t, err)
	assert.Equal(t, backup.ID, found.ID)
	assert.Equal(t, backup.ServiceID, found.ServiceID)
	assert.Equal(t, backup.Enabled, found.Enabled)
	assert.Equal(t, backup.Path, found.Path)
	assert.Equal(t, backup.Schedule, found.Schedule)
	assert.Equal(t, backup.Status, found.Status)
}

func TestBackupRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
}

func TestBackupRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)

	_, backup := createTestBackup(t, db, repo, "update")
	backup.Status = "failed"
	backup.Enabled = false
	backup.Schedule = "0 3 * * *"

	err := repo.Update(context.Background(), backup)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), backup.ID)
	assert.NoError(t, err)
	assert.Equal(t, "failed", found.Status)
	assert.False(t, found.Enabled)
	assert.Equal(t, "0 3 * * *", found.Schedule)
}

func TestBackupRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)

	_, backup := createTestBackup(t, db, repo, "delete")

	err := repo.Delete(context.Background(), backup.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), backup.ID)
	assert.Error(t, err)
}

func TestBackupRepository_ListByService(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)

	svc, _ := createTestBackup(t, db, repo, "list1")
	_, _ = createTestBackup(t, db, repo, "list2")

	// Add another backup for the first service
	now := time.Now()
	backup3 := &model.Backup{
		ServiceID:      svc.ID,
		Enabled:        true,
		Path:           "/backups/extra",
		Schedule:       "0 4 * * *",
		LastBackupTime: &now,
		Status:         "success",
	}
	require.NoError(t, repo.Create(context.Background(), backup3))

	backups, total, err := repo.ListByService(context.Background(), svc.ID, &model.ListBackupsRequest{
		Page:     1,
		PageSize: 20,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, backups, 2)
}

func TestBackupRepository_ListByService_WithStatusFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)

	svc, _ := createTestBackup(t, db, repo, "filter") // status=success

	// Create a failed backup
	now := time.Now()
	failedBackup := &model.Backup{
		ServiceID:      svc.ID,
		Enabled:        true,
		Path:           "/backups/failed",
		Schedule:       "0 2 * * *",
		LastBackupTime: &now,
		Status:         "failed",
	}
	require.NoError(t, repo.Create(context.Background(), failedBackup))

	backups, total, err := repo.ListByService(context.Background(), svc.ID, &model.ListBackupsRequest{
		Page:     1,
		PageSize: 20,
		Status:   "success",
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, backups, 1)
	assert.Equal(t, "success", backups[0].Status)
}

func TestBackupRepository_ListByService_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)

	svc, _ := createTestBackup(t, db, repo, "page")

	for i := 0; i < 4; i++ {
		now := time.Now()
		b := &model.Backup{
			ServiceID:      svc.ID,
			Enabled:        true,
			Path:           fmt.Sprintf("/backups/pg_%d", i),
			Schedule:       "0 2 * * *",
			LastBackupTime: &now,
			Status:         "success",
		}
		require.NoError(t, repo.Create(context.Background(), b))
	}

	backups, total, err := repo.ListByService(context.Background(), svc.ID, &model.ListBackupsRequest{
		Page:     1,
		PageSize: 2,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total) // 1 from helper + 4
	assert.Len(t, backups, 2)
}

func TestNewBackupRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)
	assert.NotNil(t, repo)
}
