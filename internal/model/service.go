package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EnvironmentID  uuid.UUID `gorm:"type:uuid;not null;index" json:"environment_id"`
	Environment    *Environment `gorm:"foreignKey:EnvironmentID" json:"environment,omitempty"`
	ServerID       uuid.UUID `gorm:"type:uuid;not null;index" json:"server_id"`
	Server         *Server   `gorm:"foreignKey:ServerID" json:"server,omitempty"`
	Name           string    `gorm:"not null;size:255" json:"name"`
	Type           string    `gorm:"size:100" json:"type"`
	URL            string    `gorm:"size:500" json:"url"`
	Repository     string    `gorm:"size:500" json:"repository"`
	StackLanguage  string    `gorm:"size:100" json:"stack_language"`
	StackFramework string    `gorm:"size:100" json:"stack_framework"`
	DBType         string    `gorm:"size:100" json:"db_type"`
	DBHost         string    `gorm:"size:500" json:"db_host"`
	CreatedAt         time.Time           `gorm:"autoCreateTime" json:"created_at"`

	MonitoringConfig  *MonitoringConfig   `gorm:"foreignKey:ServiceID" json:"monitoring_config,omitempty"`
	MonitoringLogs    []MonitoringLog     `gorm:"foreignKey:ServiceID" json:"monitoring_logs,omitempty"`
	Deployments       []Deployment        `gorm:"foreignKey:ServiceID" json:"deployments,omitempty"`
	Backups           []Backup            `gorm:"foreignKey:ServiceID" json:"backups,omitempty"`
}

func (s *Service) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type CreateServiceRequest struct {
	EnvironmentID  uuid.UUID `json:"environment_id" binding:"required"`
	ServerID       uuid.UUID `json:"server_id" binding:"required"`
	Name           string    `json:"name" binding:"required,min=1,max=255"`
	Type           string    `json:"type" binding:"omitempty,max=100"`
	URL            string    `json:"url" binding:"omitempty,max=500"`
	Repository     string    `json:"repository" binding:"omitempty,max=500"`
	StackLanguage  string    `json:"stack_language" binding:"omitempty,max=100"`
	StackFramework string    `json:"stack_framework" binding:"omitempty,max=100"`
	DBType         string    `json:"db_type" binding:"omitempty,max=100"`
	DBHost         string    `json:"db_host" binding:"omitempty,max=500"`
}

type UpdateServiceRequest struct {
	EnvironmentID  *uuid.UUID `json:"environment_id"`
	ServerID       *uuid.UUID `json:"server_id"`
	Name           string     `json:"name" binding:"omitempty,min=1,max=255"`
	Type           string     `json:"type" binding:"omitempty,max=100"`
	URL            string     `json:"url" binding:"omitempty,max=500"`
	Repository     string     `json:"repository" binding:"omitempty,max=500"`
	StackLanguage  string     `json:"stack_language" binding:"omitempty,max=100"`
	StackFramework string     `json:"stack_framework" binding:"omitempty,max=100"`
	DBType         string     `json:"db_type" binding:"omitempty,max=100"`
	DBHost         string     `json:"db_host" binding:"omitempty,max=500"`
}

type ListServicesRequest struct {
	Page          int    `form:"page" binding:"omitempty,min=1"`
	PageSize      int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	EnvironmentID string `form:"environment_id" binding:"omitempty"`
	ServerID      string `form:"server_id" binding:"omitempty"`
	Search        string `form:"search" binding:"omitempty,max=255"`
}
