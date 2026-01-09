package model

import (
	"time"
)

// ProjectKey API 密钥
type ProjectKey struct {
	ID            int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	ProjectID     int64      `json:"project_id" gorm:"index;not null"`
	Name          string     `json:"name" gorm:"type:varchar(100)"`
	AccessKey     string     `json:"access_key" gorm:"type:varchar(64);uniqueIndex;not null"`
	SecretKeyHash string     `json:"-" gorm:"type:varchar(128);not null"`
	Permissions   string     `json:"permissions" gorm:"type:json"` // {"read": true, "write": false, ...}
	IPWhitelist   string     `json:"ip_whitelist,omitempty" gorm:"type:json"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 表名
func (ProjectKey) TableName() string {
	return "project_keys"
}

// Permissions 权限结构
type Permissions struct {
	Read    bool `json:"read"`
	Write   bool `json:"write"`
	Delete  bool `json:"delete"`
	Release bool `json:"release"`
	Admin   bool `json:"admin"`
	Decrypt bool `json:"decrypt"`
}

// DefaultPermissions 默认权限
func DefaultPermissions() Permissions {
	return Permissions{
		Read:    true,
		Write:   true,
		Delete:  false,
		Release: false,
		Admin:   false,
		Decrypt: false,
	}
}
