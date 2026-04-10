package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Deployment struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ServiceID     uuid.UUID `gorm:"type:uuid;not null;index" json:"service_id"`
	Service       *Service  `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Method        string    `gorm:"not null;size:100" json:"method"`
	ContainerName string    `gorm:"size:255" json:"container_name"`
	Port          int       `json:"port"`
	Config        string    `gorm:"type:jsonb" json:"config"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (d *Deployment) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

type CreateDeploymentRequest struct {
	ServiceID     uuid.UUID `json:"service_id" binding:"required"`
	Method        string    `json:"method" binding:"required,min=1,max=100"`
	ContainerName string    `json:"container_name" binding:"omitempty,max=255"`
	Port          int       `json:"port" binding:"omitempty,min=1,max=65535"`
	Config        string    `json:"config" binding:"omitempty"`
}

type UpdateDeploymentRequest struct {
	Method        string `json:"method" binding:"omitempty,min=1,max=100"`
	ContainerName string `json:"container_name" binding:"omitempty,max=255"`
	Port          int    `json:"port" binding:"omitempty,min=1,max=65535"`
	Config        string `json:"config" binding:"omitempty"`
}

type ListDeploymentsRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	ServiceID string `form:"service_id" binding:"omitempty"`
}
