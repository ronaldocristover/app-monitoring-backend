package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MonitoringLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ServiceID      uuid.UUID `gorm:"type:uuid;not null;index" json:"service_id"`
	Service        *Service  `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Status         string    `gorm:"not null;size:20" json:"status"`
	ResponseTimeMs int       `json:"response_time_ms"`
	StatusCode     int       `json:"status_code"`
	ErrorMessage   string    `gorm:"size:1000" json:"error_message"`
	CheckedAt      time.Time `gorm:"not null;index" json:"checked_at"`
}

func (m *MonitoringLog) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

type ListMonitoringLogsRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=up down"`
}
