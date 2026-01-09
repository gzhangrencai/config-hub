package api

import (
	"net/http"
	"strconv"

	"confighub/internal/model"
	"confighub/internal/service"

	"github.com/gin-gonic/gin"
)

// ReleaseHandler 发布处理器
type ReleaseHandler struct {
	releaseSvc     *service.ReleaseService
	grayReleaseSvc *service.GrayReleaseService
	auditSvc       *service.AuditService
}

// NewReleaseHandler 创建发布处理器
func NewReleaseHandler(releaseSvc *service.ReleaseService, grayReleaseSvc *service.GrayReleaseService, auditSvc *service.AuditService) *ReleaseHandler {
	return &ReleaseHandler{
		releaseSvc:     releaseSvc,
		grayReleaseSvc: grayReleaseSvc,
		auditSvc:       auditSvc,
	}
}

// Create 创建发布
// POST /api/configs/:id/release
func (h *ReleaseHandler) Create(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	var req struct {
		Environment string `json:"environment" binding:"required"`
		Version     int    `json:"version"`
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

	release, err := h.releaseSvc.Create(c.Request.Context(), configID, req.Environment, req.Version, author)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 记录审计日志
	h.auditSvc.Log(c.Request.Context(), &model.AuditLog{
		ProjectID:    release.ProjectID,
		UserID:       &userID,
		Action:       model.AuditActionRelease,
		ResourceType: model.AuditResourceRelease,
		ResourceID:   release.ID,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusCreated, gin.H{
		"release": release,
	})
}

// List 获取发布历史
// GET /api/configs/:id/releases
func (h *ReleaseHandler) List(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	releases, err := h.releaseSvc.List(c.Request.Context(), configID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"releases": releases,
	})
}


// Rollback 回滚发布
// POST /api/releases/:id/rollback
func (h *ReleaseHandler) Rollback(c *gin.Context) {
	releaseID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的发布 ID",
		})
		return
	}

	userID := getUserID(c)
	author := "user"
	if userID > 0 {
		author = strconv.FormatInt(userID, 10)
	}

	release, err := h.releaseSvc.Rollback(c.Request.Context(), releaseID, author)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"release": release,
	})
}

// CreateGray 创建灰度发布
// POST /api/configs/:id/gray-release
func (h *ReleaseHandler) CreateGray(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	var req struct {
		Environment string   `json:"environment" binding:"required"`
		Version     int      `json:"version"`
		RuleType    string   `json:"rule_type" binding:"required"` // percentage, client_id, ip_range
		Percentage  int      `json:"percentage,omitempty"`
		ClientIDs   []string `json:"client_ids,omitempty"`
		IPRanges    []string `json:"ip_ranges,omitempty"`
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

	grayReq := &service.GrayReleaseRequest{
		ConfigID:    configID,
		Environment: req.Environment,
		Version:     req.Version,
		RuleType:    req.RuleType,
		Percentage:  req.Percentage,
		ClientIDs:   req.ClientIDs,
		IPRanges:    req.IPRanges,
	}

	release, err := h.grayReleaseSvc.Create(c.Request.Context(), grayReq, author)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 记录审计日志
	h.auditSvc.Log(c.Request.Context(), &model.AuditLog{
		ProjectID:    release.ProjectID,
		UserID:       &userID,
		Action:       "gray_release",
		ResourceType: model.AuditResourceRelease,
		ResourceID:   release.ID,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusCreated, gin.H{
		"release": release,
	})
}

// Promote 提升灰度发布为正式发布
// POST /api/releases/:id/promote
func (h *ReleaseHandler) Promote(c *gin.Context) {
	releaseID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的发布 ID",
		})
		return
	}

	userID := getUserID(c)
	author := "user"
	if userID > 0 {
		author = strconv.FormatInt(userID, 10)
	}

	release, err := h.grayReleaseSvc.Promote(c.Request.Context(), releaseID, author)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "灰度发布已提升为正式发布",
		"release": release,
	})
}

// Cancel 取消灰度发布
// POST /api/releases/:id/cancel
func (h *ReleaseHandler) Cancel(c *gin.Context) {
	releaseID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的发布 ID",
		})
		return
	}

	if err := h.grayReleaseSvc.Cancel(c.Request.Context(), releaseID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "灰度发布已取消",
	})
}

// UpdateGrayPercentage 更新灰度百分比
// PUT /api/releases/:id/percentage
func (h *ReleaseHandler) UpdateGrayPercentage(c *gin.Context) {
	releaseID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的发布 ID",
		})
		return
	}

	var req struct {
		Percentage int `json:"percentage" binding:"required,min=0,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	if err := h.grayReleaseSvc.UpdatePercentage(c.Request.Context(), releaseID, req.Percentage); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "灰度百分比已更新",
	})
}

// ListEnvironments 获取环境列表
// GET /api/projects/:id/environments
func (h *ReleaseHandler) ListEnvironments(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的项目 ID",
		})
		return
	}

	envs, err := h.releaseSvc.ListEnvironments(c.Request.Context(), projectID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"environments": envs,
	})
}

// CreateEnvironment 创建环境
// POST /api/projects/:id/environments
func (h *ReleaseHandler) CreateEnvironment(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的项目 ID",
		})
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	if err := h.releaseSvc.CreateEnvironment(c.Request.Context(), projectID, req.Name, req.Description); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "环境创建成功",
	})
}
