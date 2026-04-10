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

func createTestServiceForBackup(t *testing.T, db *gorm.DB) *model.Service {
	t.Helper()
	app := &model.App{AppName: "BackupApp_" + t.Name()}
	require.NoError(t, db.Create(app).Error)
	server := &model.Server{Name: "BackupServer_" + t.Name(), IP: "10.0.0.4"}
	require.NoError(t, db.Create(server).Error)
	env := &model.Environment{AppID: app.ID, Name: "production_" + t.Name()}
	require.NoError(t, db.Create(env).Error)
	svc := &model.Service{
		EnvironmentID: env.ID,
		ServerID:      server.ID,
		Name:          "BackupService_" + t.Name(),
	}
	require.NoError(t, db.Create(svc).Error)
	return svc
}

func TestBackupRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)
	ctx := context.Background()

	svc := createTestServiceForBackup(t, db)

	backup := &model.Backup{
		ServiceID: svc.ID,
		Enabled:   true,
		Path:      "/backups/test",
		Schedule:  "0 2 * * *",
		Status:    "completed",
	}

	err := repo.Create(ctx, backup)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, backup.ID)
}

func TestBackupRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)
	ctx := context.Background()

	svc := createTestServiceForBackup(t, db)

	backup := &model.Backup{
		ServiceID: svc.ID,
		Enabled:   true,
		Path:      "/backups/test",
		Schedule:  "0 2 * * *",
		Status:    "completed",
	}
	require.NoError(t, repo.Create(ctx, backup))

	found, err := repo.GetByID(ctx, backup.ID)
	assert.NoError(t, err)
	assert.Equal(t, backup.ID, found.ID)
	assert.Equal(t, svc.ID, found.ServiceID)
	assert.Equal(t, "/backups/test", found.Path)
	assert.Equal(t, "completed", found.Status)
	assert.True(t, found.Enabled)

	// Test non-existent ID
	_, err = repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
}

func TestBackupRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)
	ctx := context.Background()

	svc := createTestServiceForBackup(t, db)

	backup := &model.Backup{
		ServiceID: svc.ID,
		Enabled:   true,
		Path:      "/backups/original",
		Schedule:  "0 2 * * *",
		Status:    "completed",
	}
	require.NoError(t, repo.Create(ctx, backup))

	backup.Path = "/backups/updated"
	backup.Status = "failed"
	err := repo.Update(ctx, backup)
	assert.NoError(t, err)

	found, err := repo.GetByID(ctx, backup.ID)
	assert.NoError(t, err)
	assert.Equal(t, "/backups/updated", found.Path)
	assert.Equal(t, "failed", found.Status)
}

func TestBackupRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)
	ctx := context.Background()

	svc := createTestServiceForBackup(t, db)

	backup := &model.Backup{
		ServiceID: svc.ID,
		Enabled:   true,
		Path:      "/backups/delete-test",
		Schedule:  "0 2 * * *",
		Status:    "completed",
	}
	require.NoError(t, repo.Create(ctx, backup))

	err := repo.Delete(ctx, backup.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(ctx, backup.ID)
	assert.Error(t, err)
}

func TestBackupRepository_ListByService(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)
	ctx := context.Background()

	svc := createTestServiceForBackup(t, db)

	// Count before inserting
	var countBefore int64
	db.Model(&model.Backup{}).Where("service_id = ?", svc.ID).Count(&countBefore)

	// Create 3 backups for the same service
	now := time.Now()
	for i := 0; i < 3; i++ {
		backup := &model.Backup{
			ServiceID:      svc.ID,
			Enabled:        true,
			Path:           fmt.Sprintf("/backups/backup_%d", i),
			Schedule:       "0 2 * * *",
			LastBackupTime: &now,
			Status:         "completed",
		}
		require.NoError(t, repo.Create(ctx, backup))
	}

	backups, total, err := repo.ListByService(ctx, svc.ID, &model.ListBackupsRequest{
		Page:     1,
		PageSize: 10,
	})
	assert.NoError(t, err)
	assert.Equal(t, countBefore+3, total)
	assert.Len(t, backups, 3)
}

func TestBackupRepository_ListByService_WithStatusFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewBackupRepository(db)
	ctx := context.Background()

	svc := createTestServiceForBackup(t, db)

	now := time.Now()
	// Create 2 backups with status="completed"
	require.NoError(t, repo.Create(ctx, &model.Backup{
		ServiceID:      svc.ID,
		Enabled:        true,
		Path:           "/backups/completed_1",
		Schedule:       "0 2 * * *",
		LastBackupTime: &now,
		Status:         "completed",
	}))
	require.NoError(t, repo.Create(ctx, &model.Backup{
		ServiceID:      svc.ID,
		Enabled:        true,
		Path:           "/backups/completed_2",
		Schedule:       "0 2 * * *",
		LastBackupTime: &now,
		Status:         "completed",
	}))
	// Create 1 backup with status="failed"
	require.NoError(t, repo.Create(ctx, &model.Backup{
		ServiceID:      svc.ID,
		Enabled:        false,
		Path:           "/backups/failed_1",
		Schedule:       "0 2 * * *",
		LastBackupTime: &now,
		Status:         "failed",
	}))

	backups, total, err := repo.ListByService(ctx, svc.ID, &model.ListBackupsRequest{
		Page:     1,
		PageSize: 10,
		Status:   "completed",
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, backups, 2)
	for _, b := range backups {
		assert.Equal(t, "completed", b.Status)
	}
}
