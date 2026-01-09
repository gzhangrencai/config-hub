package model

import (
	"time"
)

// Project 项目
type Project struct {
	ID                int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name              string    `json:"name" gorm:"type:varchar(100);uniqueIndex;not null"`
	Description       string    `json:"description" gorm:"type:text"`
	AccessMode        string    `json:"access_mode" gorm:"type:varchar(20);default:key"` // public, key, auth
	PublicPermissions string    `json:"public_permissions" gorm:"type:json"`
	Settings          string    `json:"settings,omitempty" gorm:"type:json"`
	GitRepoURL        string    `json:"git_repo_url,omitempty" gorm:"type:varchar(500)"`
	GitBranch         string    `json:"git_branch,omitempty" gorm:"type:varchar(100);default:main"`
	WebhookSecret     string    `json:"-" gorm:"type:varchar(128)"`
	CreatedBy         int64     `json:"created_by" gorm:"index"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 表名
func (Project) TableName() string {
	return "projects"
}

// ProjectEnvironment 项目环境
type ProjectEnvironment struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ProjectID   int64     `json:"project_id" gorm:"index;not null"`
	Name        string    `json:"name" gorm:"type:varchar(50);not null"`
	Description string    `json:"description" gorm:"type:varchar(200)"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 表名
func (ProjectEnvironment) TableName() string {
	return "project_environments"
}
