package api

import (
	"net/http"
	"strconv"

	"confighub/internal/model"
	"confighub/internal/service"

	"github.com/gin-gonic/gin"
)

// KeyHandler 密钥处理器
type KeyHandler struct {
	keySvc   *service.KeyService
	auditSvc *service.AuditService
}

// NewKeyHandler 创建密钥处理器
func NewKeyHandler(keySvc *service.KeyService, auditSvc *service.AuditService) *KeyHandler {
	return &KeyHandler{
		keySvc:   keySvc,
		auditSvc: auditSvc,
	}
}

// Create 创建密钥
// POST /api/projects/:id/keys
func (h *KeyHandler) Create(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的项目 ID",
		})
		return
	}

	var req service.CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	key, secretKey, err := h.keySvc.Create(c.Request.Context(), projectID, &req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 记录审计日志
	userID := getUserID(c)
	h.auditSvc.Log(c.Request.Context(), &model.AuditLog{
		ProjectID:    projectID,
		UserID:       &userID,
		Action:       model.AuditActionCreate,
		ResourceType: model.AuditResourceKey,
		ResourceID:   key.ID,
		ResourceName: key.Name,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusCreated, gin.H{
		"key":        key,
		"secret_key": secretKey, // 仅此一次返回
	})
}

// List 获取密钥列表
// GET /api/projects/:id/keys
func (h *KeyHandler) List(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的项目 ID",
		})
		return
	}

	keys, err := h.keySvc.List(c.Request.Context(), projectID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keys": keys,
	})
}


// Update 更新密钥
// PUT /api/keys/:id
func (h *KeyHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的密钥 ID",
		})
		return
	}

	var req service.UpdateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	if err := h.keySvc.Update(c.Request.Context(), id, &req); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
	})
}

// Delete 删除密钥
// DELETE /api/keys/:id
func (h *KeyHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的密钥 ID",
		})
		return
	}

	if err := h.keySvc.Delete(c.Request.Context(), id); err != nil {
		handleServiceError(c, err)
		return
	}

	// 记录审计日志
	userID := getUserID(c)
	h.auditSvc.Log(c.Request.Context(), &model.AuditLog{
		UserID:       &userID,
		Action:       model.AuditActionDelete,
		ResourceType: model.AuditResourceKey,
		ResourceID:   id,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "删除成功",
	})
}

// Regenerate 重新生成密钥
// POST /api/keys/:id/regenerate
func (h *KeyHandler) Regenerate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的密钥 ID",
		})
		return
	}

	key, secretKey, err := h.keySvc.Regenerate(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":        key,
		"secret_key": secretKey,
	})
}
