package api

import (
	"net/http"

	"confighub/internal/middleware"
	"confighub/internal/service"

	"github.com/gin-gonic/gin"
)

// getUserID 从上下文获取用户 ID
func getUserID(c *gin.Context) int64 {
	if authCtx, exists := c.Get("auth_context"); exists {
		if ctx, ok := authCtx.(*middleware.AuthContext); ok {
			return ctx.UserID
		}
	}
	return 0
}

// getAccessKeyID 从上下文获取 Access Key ID
func getAccessKeyID(c *gin.Context) int64 {
	if authCtx, exists := c.Get("auth_context"); exists {
		if ctx, ok := authCtx.(*middleware.AuthContext); ok {
			return ctx.AccessKeyID
		}
	}
	return 0
}

// getProjectID 从上下文获取项目 ID
func getProjectID(c *gin.Context) int64 {
	if authCtx, exists := c.Get("auth_context"); exists {
		if ctx, ok := authCtx.(*middleware.AuthContext); ok {
			return ctx.ProjectID
		}
	}
	return 0
}

// handleServiceError 处理服务层错误
func handleServiceError(c *gin.Context, err error) {
	switch err {
	case service.ErrProjectNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "项目不存在",
		})
	case service.ErrProjectNameExists:
		c.JSON(http.StatusConflict, gin.H{
			"code":    "CONFLICT",
			"message": "项目名称已存在",
		})
	case service.ErrProjectNameRequired:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "项目名称不能为空",
		})
	case service.ErrConfigNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "配置不存在",
		})
	case service.ErrConfigNameExists:
		c.JSON(http.StatusConflict, gin.H{
			"code":    "CONFLICT",
			"message": "配置名称已存在",
		})
	case service.ErrInvalidJSON:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": "无效的 JSON 格式",
		})
	case service.ErrInvalidYAML:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": "无效的 YAML 格式",
		})
	case service.ErrInvalidFileType:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": "不支持的文件类型",
		})
	case service.ErrVersionNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "版本不存在",
		})
	case service.ErrKeyNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "密钥不存在",
		})
	case service.ErrReleaseNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "发布记录不存在",
		})
	case service.ErrInvalidSchema:
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "VALIDATION_ERROR",
			"message": "无效的 Schema 定义",
		})
	case service.ErrSchemaNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "Schema 不存在",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "服务器内部错误",
			"details": err.Error(),
		})
	}
}
