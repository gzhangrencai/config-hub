package model

import (
	"time"
)

// Release 发布记录
type Release struct {
	ID             int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ProjectID      int64     `json:"project_id" gorm:"index;not null"`
	ConfigID       int64     `json:"config_id" gorm:"index;not null"`
	Version        int       `json:"version" gorm:"not null"`
	Environment    string    `json:"environment" gorm:"type:varchar(50);not null"`
	Status         string    `json:"status" gorm:"type:varchar(20);default:released"` // pending, released, rollback, gray
	ReleaseType    string    `json:"release_type" gorm:"type:varchar(10);default:full"` // full, gray
	GrayRules      string    `json:"gray_rules,omitempty" gorm:"type:json"`
	GrayPercentage int       `json:"gray_percentage,omitempty" gorm:"default:0"`
	ReleasedBy     string    `json:"released_by" gorm:"type:varchar(100)"`
	ReleasedAt     time.Time `json:"released_at" gorm:"autoCreateTime"`
}

// TableName 表名
func (Release) TableName() string {
	return "releases"
}

// GrayRules 灰度发布规则
type GrayRules struct {
	Type       string   `json:"type"`                  // percentage, client_id, ip_range
	Percentage int      `json:"percentage,omitempty"`
	ClientIDs  []string `json:"client_ids,omitempty"`
	IPRanges   []string `json:"ip_ranges,omitempty"`
}
