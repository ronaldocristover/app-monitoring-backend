package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Environment struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppID     uuid.UUID `gorm:"type:uuid;not null;index" json:"app_id"`
	App       *App      `gorm:"foreignKey:AppID" json:"app,omitempty"`
	Name      string    `gorm:"not null;size:255" json:"name"`
	Services  []Service `gorm:"foreignKey:EnvironmentID" json:"services,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (e *Environment) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

type CreateEnvironmentRequest struct {
	AppID uuid.UUID `json:"app_id" binding:"required"`
	Name  string    `json:"name" binding:"required,min=1,max=255"`
}

type UpdateEnvironmentRequest struct {
	Name string `json:"name" binding:"omitempty,min=1,max=255"`
}

type ListEnvironmentsRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	AppID    string `form:"app_id" binding:"omitempty"`
}
