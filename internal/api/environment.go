package api

import (
	"net/http"
	"strconv"

	"confighub/internal/service"

	"github.com/gin-gonic/gin"
)

// EnvironmentHandler 环境处理器
type EnvironmentHandler struct {
	envSvc     *service.EnvironmentService
	envDiffSvc *service.EnvDiffService
}

// NewEnvironmentHandler 创建环境处理器
func NewEnvironmentHandler(envSvc *service.EnvironmentService, envDiffSvc *service.EnvDiffService) *EnvironmentHandler {
	return &EnvironmentHandler{
		envSvc:     envSvc,
		envDiffSvc: envDiffSvc,
	}
}

// List 获取环境列表
// GET /api/projects/:id/environments
func (h *EnvironmentHandler) List(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的项目 ID",
		})
		return
	}

	envs, err := h.envSvc.List(c.Request.Context(), projectID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"environments": envs,
	})
}

// Create 创建环境
// POST /api/projects/:id/environments
func (h *EnvironmentHandler) Create(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的项目 ID",
		})
		return
	}

	var req service.Environment
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	if err := h.envSvc.Create(c.Request.Context(), projectID, req); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "环境创建成功",
	})
}

// Compare 对比环境配置
// GET /api/configs/:id/compare?source=dev&target=prod
func (h *EnvironmentHandler) Compare(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	sourceEnv := c.Query("source")
	targetEnv := c.Query("target")

	if sourceEnv == "" || targetEnv == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请指定源环境和目标环境",
		})
		return
	}

	comparison, err := h.envDiffSvc.Compare(c.Request.Context(), configID, sourceEnv, targetEnv)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, comparison)
}

// Sync 同步配置到目标环境
// POST /api/configs/:id/sync
func (h *EnvironmentHandler) Sync(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	var req struct {
		SourceEnv string   `json:"source_env" binding:"required"`
		TargetEnv string   `json:"target_env" binding:"required"`
		Keys      []string `json:"keys"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	if err := h.envDiffSvc.Sync(c.Request.Context(), configID, req.SourceEnv, req.TargetEnv, req.Keys); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "同步成功",
	})
}

// MergeConfig 合并配置
// POST /api/configs/:id/merge
func (h *EnvironmentHandler) MergeConfig(c *gin.Context) {
	var req struct {
		BaseContent string `json:"base_content" binding:"required"`
		EnvContent  string `json:"env_content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	merged, err := h.envSvc.MergeConfig(c.Request.Context(), req.BaseContent, req.EnvContent)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": merged,
	})
}

