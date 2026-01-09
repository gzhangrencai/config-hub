package api

import (
	"net/http"
	"strconv"

	"confighub/internal/service"

	"github.com/gin-gonic/gin"
)

// SchemaHandler Schema 处理器
type SchemaHandler struct {
	schemaSvc *service.SchemaService
}

// NewSchemaHandler 创建 Schema 处理器
func NewSchemaHandler(schemaSvc *service.SchemaService) *SchemaHandler {
	return &SchemaHandler{
		schemaSvc: schemaSvc,
	}
}

// Get 获取 Schema
// GET /api/configs/:id/schema
func (h *SchemaHandler) Get(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	schema, err := h.schemaSvc.Get(c.Request.Context(), configID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"schema": schema,
	})
}

// Update 更新 Schema
// PUT /api/configs/:id/schema
func (h *SchemaHandler) Update(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	var req struct {
		Schema string `json:"schema" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	if err := h.schemaSvc.Update(c.Request.Context(), configID, req.Schema); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Schema 更新成功",
	})
}

// Generate 自动生成 Schema
// POST /api/configs/:id/schema/generate
func (h *SchemaHandler) Generate(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	schema, err := h.schemaSvc.Generate(c.Request.Context(), configID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"schema": schema,
	})
}
