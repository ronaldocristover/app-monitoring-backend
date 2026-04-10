package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database with all models migrated
// using raw SQL to avoid gen_random_uuid() which is PostgreSQL-specific.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Ensure single connection so :memory: data is shared across all operations
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS apps (
			id TEXT PRIMARY KEY,
			app_name TEXT NOT NULL,
			description TEXT,
			tags TEXT,
			created_at DATETIME
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS environments (
			id TEXT PRIMARY KEY,
			app_id TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at DATETIME,
			FOREIGN KEY (app_id) REFERENCES apps(id)
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS servers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			ip TEXT NOT NULL,
			provider TEXT,
			created_at DATETIME
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS services (
			id TEXT PRIMARY KEY,
			environment_id TEXT NOT NULL,
			server_id TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT,
			url TEXT,
			repository TEXT,
			stack_language TEXT,
			stack_framework TEXT,
			db_type TEXT,
			db_host TEXT,
			created_at DATETIME,
			FOREIGN KEY (environment_id) REFERENCES environments(id),
			FOREIGN KEY (server_id) REFERENCES servers(id)
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS monitoring_configs (
			id TEXT PRIMARY KEY,
			service_id TEXT NOT NULL UNIQUE,
			enabled BOOLEAN DEFAULT TRUE,
			ping_interval_seconds INTEGER DEFAULT 60,
			timeout_seconds INTEGER DEFAULT 10,
			retries INTEGER DEFAULT 3,
			FOREIGN KEY (service_id) REFERENCES services(id)
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS monitoring_logs (
			id TEXT PRIMARY KEY,
			service_id TEXT NOT NULL,
			status TEXT NOT NULL,
			response_time_ms INTEGER,
			status_code INTEGER,
			error_message TEXT,
			checked_at DATETIME NOT NULL,
			FOREIGN KEY (service_id) REFERENCES services(id)
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS deployments (
			id TEXT PRIMARY KEY,
			service_id TEXT NOT NULL,
			method TEXT NOT NULL,
			container_name TEXT,
			port INTEGER,
			config TEXT,
			created_at DATETIME,
			FOREIGN KEY (service_id) REFERENCES services(id)
		)
	`).Error)

	require.NoError(t, db.Exec(`
		CREATE TABLE IF NOT EXISTS backups (
			id TEXT PRIMARY KEY,
			service_id TEXT NOT NULL,
			enabled BOOLEAN DEFAULT FALSE,
			path TEXT,
			schedule TEXT,
			last_backup_time DATETIME,
			status TEXT,
			FOREIGN KEY (service_id) REFERENCES services(id)
		)
	`).Error)

	return db
}
