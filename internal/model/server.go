package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Server struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"not null;size:255" json:"name"`
	IP        string    `gorm:"not null;size:45" json:"ip"`
	Provider  string    `gorm:"size:100" json:"provider"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (s *Server) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type CreateServerRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=255"`
	IP       string `json:"ip" binding:"required,min=1,max=45"`
	Provider string `json:"provider" binding:"omitempty,max=100"`
}

type UpdateServerRequest struct {
	Name     string `json:"name" binding:"omitempty,min=1,max=255"`
	IP       string `json:"ip" binding:"omitempty,min=1,max=45"`
	Provider string `json:"provider" binding:"omitempty,max=100"`
}

type ListServersRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search   string `form:"search" binding:"omitempty,max=255"`
}
