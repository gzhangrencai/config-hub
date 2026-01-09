package model

import (
	"time"
)

// User 用户
type User struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string    `json:"username" gorm:"type:varchar(100);uniqueIndex;not null"`
	Email        string    `json:"email" gorm:"type:varchar(200);uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"type:varchar(128);not null"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// ProjectMember 项目成员
type ProjectMember struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ProjectID int64     `json:"project_id" gorm:"index;not null"`
	UserID    int64     `json:"user_id" gorm:"index;not null"`
	Role      string    `json:"role" gorm:"type:varchar(20);default:viewer"` // viewer, developer, releaser, admin
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 表名
func (ProjectMember) TableName() string {
	return "project_members"
}

// ClientConnection 客户端连接 (用于实时推送)
type ClientConnection struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ClientID      string    `json:"client_id" gorm:"type:varchar(100);index;not null"`
	ProjectID     int64     `json:"project_id" gorm:"index;not null"`
	ConfigIDs     string    `json:"config_ids" gorm:"type:json"`
	LastVersion   string    `json:"last_version" gorm:"type:json"`
	IPAddress     string    `json:"ip_address" gorm:"type:varchar(45)"`
	ConnectedAt   time.Time `json:"connected_at" gorm:"autoCreateTime"`
	LastHeartbeat time.Time `json:"last_heartbeat" gorm:"index;autoUpdateTime"`
}

// TableName 表名
func (ClientConnection) TableName() string {
	return "client_connections"
}
