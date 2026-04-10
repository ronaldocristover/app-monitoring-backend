package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Backup struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ServiceID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"service_id"`
	Service        *Service   `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Enabled        bool       `gorm:"default:false" json:"enabled"`
	Path           string     `gorm:"size:500" json:"path"`
	Schedule       string     `gorm:"size:100" json:"schedule"`
	LastBackupTime *time.Time `json:"last_backup_time"`
	Status         string     `gorm:"size:50" json:"status"`
}

func (b *Backup) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

type CreateBackupRequest struct {
	ServiceID uuid.UUID `json:"service_id" binding:"required"`
	Enabled   bool      `json:"enabled"`
	Path      string    `json:"path" binding:"omitempty,max=500"`
	Schedule  string    `json:"schedule" binding:"omitempty,max=100"`
}

type UpdateBackupRequest struct {
	Enabled   *bool   `json:"enabled"`
	Path      string  `json:"path" binding:"omitempty,max=500"`
	Schedule  string  `json:"schedule" binding:"omitempty,max=100"`
	Status    string  `json:"status" binding:"omitempty,max=50"`
}

type ListBackupsRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	ServiceID string `form:"service_id" binding:"omitempty"`
	Status    string `form:"status" binding:"omitempty,max=50"`
}
