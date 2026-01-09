package api

import (
	"net/http"
	"strconv"
	"time"

	"confighub/internal/repository"
	"confighub/internal/service"

	"github.com/gin-gonic/gin"
)

// AuditHandler 审计日志处理器
type AuditHandler struct {
	auditSvc *service.AuditService
}

// NewAuditHandler 创建审计日志处理器
func NewAuditHandler(auditSvc *service.AuditService) *AuditHandler {
	return &AuditHandler{
		auditSvc: auditSvc,
	}
}

// List 获取审计日志列表
// GET /api/projects/:id/audit-logs
func (h *AuditHandler) List(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的项目 ID",
		})
		return
	}

	// 解析查询参数
	filter := &repository.AuditFilter{
		ProjectID: projectID,
		Action:    c.Query("action"),
		Limit:     100,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartTime = &t
		}
	}

	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndTime = &t
		}
	}

	logs, err := h.auditSvc.List(c.Request.Context(), filter)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
	})
}
