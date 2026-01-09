package repository

import (
	"context"
	"time"

	"confighub/internal/model"

	"gorm.io/gorm"
)

// AuditRepository 审计日志数据访问
type AuditRepository struct {
	db *gorm.DB
}

// NewAuditRepository 创建审计日志仓库
func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create 创建审计日志
func (r *AuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// AuditFilter 审计日志过滤条件
type AuditFilter struct {
	ProjectID    int64
	Action       string
	ResourceType string
	StartTime    *time.Time
	EndTime      *time.Time
	Limit        int
	Offset       int
}

// List 获取审计日志列表
func (r *AuditRepository) List(ctx context.Context, filter *AuditFilter) ([]*model.AuditLog, error) {
	var logs []*model.AuditLog
	query := r.db.WithContext(ctx).Model(&model.AuditLog{})

	if filter.ProjectID > 0 {
		query = query.Where("project_id = ?", filter.ProjectID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.ResourceType != "" {
		query = query.Where("resource_type = ?", filter.ResourceType)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	} else {
		query = query.Limit(100)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// Count 统计审计日志数量
func (r *AuditRepository) Count(ctx context.Context, filter *AuditFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.AuditLog{})

	if filter.ProjectID > 0 {
		query = query.Where("project_id = ?", filter.ProjectID)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.ResourceType != "" {
		query = query.Where("resource_type = ?", filter.ResourceType)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	err := query.Count(&count).Error
	return count, err
}
