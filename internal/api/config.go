package api

import (
	"net/http"
	"strconv"

	"confighub/internal/model"
	"confighub/internal/service"

	"github.com/gin-gonic/gin"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	configSvc *service.ConfigService
	auditSvc  *service.AuditService
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(configSvc *service.ConfigService, auditSvc *service.AuditService) *ConfigHandler {
	return &ConfigHandler{
		configSvc: configSvc,
		auditSvc:  auditSvc,
	}
}

// Upload 上传配置
// POST /api/projects/:id/configs
func (h *ConfigHandler) Upload(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的项目 ID",
		})
		return
	}

	var req service.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	userID := getUserID(c)
	author := "user"
	if userID > 0 {
		author = strconv.FormatInt(userID, 10)
	}

	config, err := h.configSvc.Upload(c.Request.Context(), projectID, &req, author)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 记录审计日志
	h.auditSvc.Log(c.Request.Context(), &model.AuditLog{
		ProjectID:    projectID,
		UserID:       &userID,
		Action:       model.AuditActionCreate,
		ResourceType: model.AuditResourceConfig,
		ResourceID:   config.ID,
		ResourceName: config.Name,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusCreated, gin.H{
		"config": config,
	})
}


// List 获取配置列表
// GET /api/projects/:id/configs
func (h *ConfigHandler) List(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的项目 ID",
		})
		return
	}

	configs, err := h.configSvc.List(c.Request.Context(), projectID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configs": configs,
	})
}

// Get 获取配置详情
// GET /api/configs/:id
func (h *ConfigHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	config, version, err := h.configSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"config":  config,
		"content": version.Content,
		"version": version.Version,
	})
}

// Update 更新配置
// PUT /api/configs/:id
func (h *ConfigHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	userID := getUserID(c)
	author := "user"
	if userID > 0 {
		author = strconv.FormatInt(userID, 10)
	}

	version, err := h.configSvc.Update(c.Request.Context(), id, req.Content, req.Message, author)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 记录审计日志
	config, _ := h.configSvc.GetConfigByID(c.Request.Context(), id)
	if config != nil {
		h.auditSvc.Log(c.Request.Context(), &model.AuditLog{
			ProjectID:    config.ProjectID,
			UserID:       &userID,
			Action:       model.AuditActionUpdate,
			ResourceType: model.AuditResourceConfig,
			ResourceID:   id,
			ResourceName: config.Name,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"version": version,
	})
}


// Delete 删除配置
// DELETE /api/configs/:id
func (h *ConfigHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	config, _ := h.configSvc.GetConfigByID(c.Request.Context(), id)
	if err := h.configSvc.Delete(c.Request.Context(), id); err != nil {
		handleServiceError(c, err)
		return
	}

	// 记录审计日志
	userID := getUserID(c)
	if config != nil {
		h.auditSvc.Log(c.Request.Context(), &model.AuditLog{
			ProjectID:    config.ProjectID,
			UserID:       &userID,
			Action:       model.AuditActionDelete,
			ResourceType: model.AuditResourceConfig,
			ResourceID:   id,
			ResourceName: config.Name,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// Compare 环境对比
// GET /api/configs/:id/compare
func (h *ConfigHandler) Compare(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    "NOT_IMPLEMENTED",
		"message": "环境对比功能待实现",
	})
}

// Sync 环境同步
// POST /api/configs/:id/sync
func (h *ConfigHandler) Sync(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    "NOT_IMPLEMENTED",
		"message": "环境同步功能待实现",
	})
}
