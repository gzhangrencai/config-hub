package model

import (
	"time"
)

// Config 配置文件
type Config struct {
	ID              int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ProjectID       int64     `json:"project_id" gorm:"index;not null"`
	Name            string    `json:"name" gorm:"type:varchar(200);not null"`
	Namespace       string    `json:"namespace" gorm:"type:varchar(100);default:application"`
	Environment     string    `json:"environment" gorm:"type:varchar(50);default:default"`
	FileType        string    `json:"file_type" gorm:"type:varchar(20);not null"` // json, protobuf, yaml
	SchemaJSON      string    `json:"schema_json,omitempty" gorm:"type:json"`
	DefaultEditMode string    `json:"default_edit_mode" gorm:"type:varchar(10);default:code"` // code, form
	CurrentVersion  int       `json:"current_version" gorm:"default:1"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 表名
func (Config) TableName() string {
	return "configs"
}

// ConfigVersion 配置版本
type ConfigVersion struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ConfigID      int64     `json:"config_id" gorm:"index;not null"`
	Version       int       `json:"version" gorm:"not null"`
	Content       string    `json:"content" gorm:"type:longtext"`
	CommitHash    string    `json:"commit_hash" gorm:"type:varchar(64);index"`
	CommitMessage string    `json:"commit_message" gorm:"type:varchar(500)"`
	Author        string    `json:"author" gorm:"type:varchar(100)"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 表名
func (ConfigVersion) TableName() string {
	return "config_versions"
}

// ConfigNotification 配置变更通知
type ConfigNotification struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ConfigID   int64     `json:"config_id" gorm:"index;not null"`
	Version    int       `json:"version" gorm:"not null"`
	ChangeType string    `json:"change_type" gorm:"type:varchar(20);not null"` // create, update, delete, release
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 表名
func (ConfigNotification) TableName() string {
	return "config_notifications"
}
