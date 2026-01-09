package api

import (
	"net/http"
	"strconv"

	"confighub/internal/model"
	"confighub/internal/service"

	"github.com/gin-gonic/gin"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	projectSvc *service.ProjectService
	auditSvc   *service.AuditService
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(projectSvc *service.ProjectService, auditSvc *service.AuditService) *ProjectHandler {
	return &ProjectHandler{
		projectSvc: projectSvc,
		auditSvc:   auditSvc,
	}
}

// Create 创建项目
// POST /api/projects
func (h *ProjectHandler) Create(c *gin.Context) {
	var req service.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	userID := getUserID(c)
	project, key, err := h.projectSvc.Create(c.Request.Context(), &req, userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 记录审计日志
	h.auditSvc.Log(c.Request.Context(), &model.AuditLog{
		ProjectID:    project.ID,
		UserID:       &userID,
		Action:       model.AuditActionCreate,
		ResourceType: model.AuditResourceProject,
		ResourceID:   project.ID,
		ResourceName: project.Name,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusCreated, gin.H{
		"project": project,
		"key":     key,
	})
}


// List 获取项目列表
// GET /api/projects
func (h *ProjectHandler) List(c *gin.Context) {
	userID := getUserID(c)
	projects, err := h.projectSvc.List(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
	})
}

// Get 获取项目详情
// GET /api/projects/:id
func (h *ProjectHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的项目 ID",
		})
		return
	}

	project, err := h.projectSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 获取环境列表
	envs, _ := h.projectSvc.ListEnvironments(c.Request.Context(), id)

	c.JSON(http.StatusOK, gin.H{
		"project":      project,
		"environments": envs,
	})
}

// Update 更新项目
// PUT /api/projects/:id
func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的项目 ID",
		})
		return
	}

	var req service.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	if err := h.projectSvc.Update(c.Request.Context(), id, &req); err != nil {
		handleServiceError(c, err)
		return
	}

	// 记录审计日志
	userID := getUserID(c)
	h.auditSvc.Log(c.Request.Context(), &model.AuditLog{
		ProjectID:    id,
		UserID:       &userID,
		Action:       model.AuditActionUpdate,
		ResourceType: model.AuditResourceProject,
		ResourceID:   id,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
	})
}


// Delete 删除项目
// DELETE /api/projects/:id
func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的项目 ID",
		})
		return
	}

	project, err := h.projectSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	if err := h.projectSvc.Delete(c.Request.Context(), id); err != nil {
		handleServiceError(c, err)
		return
	}

	// 记录审计日志
	userID := getUserID(c)
	h.auditSvc.Log(c.Request.Context(), &model.AuditLog{
		ProjectID:    id,
		UserID:       &userID,
		Action:       model.AuditActionDelete,
		ResourceType: model.AuditResourceProject,
		ResourceID:   id,
		ResourceName: project.Name,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// Login 用户登录 (占位)
// POST /api/auth/login
func (h *ProjectHandler) Login(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    "NOT_IMPLEMENTED",
		"message": "登录功能待实现",
	})
}

// Register 用户注册 (占位)
// POST /api/auth/register
func (h *ProjectHandler) Register(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"code":    "NOT_IMPLEMENTED",
		"message": "注册功能待实现",
	})
}
