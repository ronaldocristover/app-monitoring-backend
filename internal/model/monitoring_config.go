package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MonitoringConfig struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ServiceID           uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"service_id"`
	Service             *Service  `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Enabled             bool      `gorm:"default:true" json:"enabled"`
	PingIntervalSeconds int       `gorm:"default:60" json:"ping_interval_seconds"`
	TimeoutSeconds      int       `gorm:"default:10" json:"timeout_seconds"`
	Retries             int       `gorm:"default:3" json:"retries"`
}

func (m *MonitoringConfig) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

type UpdateMonitoringConfigRequest struct {
	Enabled             *bool `json:"enabled"`
	PingIntervalSeconds *int  `json:"ping_interval_seconds" binding:"omitempty,min=5"`
	TimeoutSeconds      *int  `json:"timeout_seconds" binding:"omitempty,min=1"`
	Retries             *int  `json:"retries" binding:"omitempty,min=0,max=10"`
}
