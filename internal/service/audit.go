package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"confighub/internal/model"
	"confighub/internal/repository"
)

// AuditService 审计日志服务
type AuditService struct {
	auditRepo *repository.AuditRepository
}

// NewAuditService 创建审计日志服务
func NewAuditService(auditRepo *repository.AuditRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

// Log 记录审计日志
func (s *AuditService) Log(ctx context.Context, entry *model.AuditLog) error {
	return s.auditRepo.Create(ctx, entry)
}

// List 获取审计日志列表
func (s *AuditService) List(ctx context.Context, filter *repository.AuditFilter) ([]*model.AuditLog, error) {
	return s.auditRepo.List(ctx, filter)
}

// Count 统计审计日志数量
func (s *AuditService) Count(ctx context.Context, filter *repository.AuditFilter) (int64, error) {
	return s.auditRepo.Count(ctx, filter)
}

// ExportCSV 导出审计日志为 CSV
func (s *AuditService) ExportCSV(ctx context.Context, filter *repository.AuditFilter, w io.Writer) error {
	logs, err := s.auditRepo.List(ctx, filter)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// 写入表头
	header := []string{"ID", "项目ID", "用户ID", "密钥ID", "操作", "资源类型", "资源ID", "资源名称", "IP地址", "用户代理", "时间"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// 写入数据
	for _, log := range logs {
		userID := ""
		if log.UserID != nil {
			userID = fmt.Sprintf("%d", *log.UserID)
		}
		keyID := ""
		if log.AccessKeyID != nil {
			keyID = fmt.Sprintf("%d", *log.AccessKeyID)
		}

		row := []string{
			fmt.Sprintf("%d", log.ID),
			fmt.Sprintf("%d", log.ProjectID),
			userID,
			keyID,
			log.Action,
			log.ResourceType,
			fmt.Sprintf("%d", log.ResourceID),
			log.ResourceName,
			log.IPAddress,
			log.UserAgent,
			log.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// ExportJSON 导出审计日志为 JSON
func (s *AuditService) ExportJSON(ctx context.Context, filter *repository.AuditFilter, w io.Writer) error {
	logs, err := s.auditRepo.List(ctx, filter)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(logs)
}

// GetStatistics 获取审计统计
func (s *AuditService) GetStatistics(ctx context.Context, projectID int64, startTime, endTime *time.Time) (*AuditStatistics, error) {
	stats := &AuditStatistics{
		ActionCounts:   make(map[string]int64),
		ResourceCounts: make(map[string]int64),
	}

	// 统计各操作类型数量
	actions := []string{model.AuditActionCreate, model.AuditActionRead, model.AuditActionUpdate, model.AuditActionDelete, model.AuditActionRelease}
	for _, action := range actions {
		filter := &repository.AuditFilter{
			ProjectID: projectID,
			Action:    action,
			StartTime: startTime,
			EndTime:   endTime,
		}
		count, err := s.auditRepo.Count(ctx, filter)
		if err != nil {
			return nil, err
		}
		stats.ActionCounts[action] = count
		stats.TotalCount += count
	}

	// 统计各资源类型数量
	resources := []string{model.AuditResourceProject, model.AuditResourceConfig, model.AuditResourceKey, model.AuditResourceRelease}
	for _, resource := range resources {
		filter := &repository.AuditFilter{
			ProjectID:    projectID,
			ResourceType: resource,
			StartTime:    startTime,
			EndTime:      endTime,
		}
		count, err := s.auditRepo.Count(ctx, filter)
		if err != nil {
			return nil, err
		}
		stats.ResourceCounts[resource] = count
	}

	return stats, nil
}

// AuditStatistics 审计统计
type AuditStatistics struct {
	TotalCount     int64            `json:"total_count"`
	ActionCounts   map[string]int64 `json:"action_counts"`
	ResourceCounts map[string]int64 `json:"resource_counts"`
}
