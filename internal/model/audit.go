package model

import (
	"time"
)

// AuditLog 审计日志
type AuditLog struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ProjectID    int64     `json:"project_id" gorm:"index"`
	UserID       *int64    `json:"user_id,omitempty"`
	AccessKeyID  *int64    `json:"access_key_id,omitempty"`
	Action       string    `json:"action" gorm:"type:varchar(50);index;not null"` // create, read, update, delete, release, login
	ResourceType string    `json:"resource_type" gorm:"type:varchar(50);not null"` // project, config, key, release
	ResourceID   int64     `json:"resource_id"`
	ResourceName string    `json:"resource_name" gorm:"type:varchar(200)"`
	IPAddress    string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent    string    `json:"user_agent" gorm:"type:varchar(500)"`
	RequestBody  string    `json:"request_body,omitempty" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"index;autoCreateTime"`
}

// TableName 表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditAction 审计动作常量
const (
	AuditActionCreate  = "create"
	AuditActionRead    = "read"
	AuditActionUpdate  = "update"
	AuditActionDelete  = "delete"
	AuditActionRelease = "release"
	AuditActionLogin   = "login"
)

// AuditResourceType 审计资源类型常量
const (
	AuditResourceProject = "project"
	AuditResourceConfig  = "config"
	AuditResourceKey     = "key"
	AuditResourceRelease = "release"
	AuditResourceUser    = "user"
)
