package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type App struct {
	ID           uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppName      string       `gorm:"not null;size:255" json:"app_name"`
	Description  string       `gorm:"size:1000" json:"description"`
	Tags         string       `gorm:"size:1000" json:"tags"`
	Environments []Environment `gorm:"foreignKey:AppID" json:"environments,omitempty"`
	CreatedAt    time.Time    `gorm:"autoCreateTime" json:"created_at"`
}

func (a *App) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type CreateAppRequest struct {
	AppName     string `json:"app_name" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"omitempty,max=1000"`
	Tags        string `json:"tags" binding:"omitempty,max=1000"`
}

type UpdateAppRequest struct {
	AppName     string `json:"app_name" binding:"omitempty,min=1,max=255"`
	Description string `json:"description" binding:"omitempty,max=1000"`
	Tags        string `json:"tags" binding:"omitempty,max=1000"`
}

type ListAppsRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search   string `form:"search" binding:"omitempty,max=255"`
	Tags     string `form:"tags" binding:"omitempty,max=1000"`
}

// Nested input structs for full create/update

type ServiceInput struct {
	ID             *uuid.UUID `json:"id"`
	ServerID       uuid.UUID  `json:"server_id" binding:"required"`
	Name           string     `json:"name" binding:"required,min=1,max=255"`
	Type           string     `json:"type" binding:"omitempty,max=100"`
	URL            string     `json:"url" binding:"omitempty,max=500"`
	Repository     string     `json:"repository" binding:"omitempty,max=500"`
	StackLanguage  string     `json:"stack_language" binding:"omitempty,max=100"`
	StackFramework string     `json:"stack_framework" binding:"omitempty,max=100"`
	DBType         string     `json:"db_type" binding:"omitempty,max=100"`
	DBHost         string     `json:"db_host" binding:"omitempty,max=500"`
}

type EnvironmentInput struct {
	ID       *uuid.UUID     `json:"id"`
	Name     string         `json:"name" binding:"required,min=1,max=255"`
	Services []ServiceInput `json:"services"`
}

type CreateFullAppRequest struct {
	AppName      string             `json:"app_name" binding:"required,min=1,max=255"`
	Description  string             `json:"description" binding:"omitempty,max=1000"`
	Tags         string             `json:"tags" binding:"omitempty,max=1000"`
	Environments []EnvironmentInput `json:"environments"`
}

type UpdateFullAppRequest struct {
	AppName      string             `json:"app_name" binding:"omitempty,min=1,max=255"`
	Description  string             `json:"description" binding:"omitempty,max=1000"`
	Tags         string             `json:"tags" binding:"omitempty,max=1000"`
	Environments []EnvironmentInput `json:"environments"`
}
