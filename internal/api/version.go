package api

import (
	"net/http"
	"strconv"

	"confighub/internal/service"

	"github.com/gin-gonic/gin"
)

// VersionHandler 版本处理器
type VersionHandler struct {
	versionSvc *service.VersionService
}

// NewVersionHandler 创建版本处理器
func NewVersionHandler(versionSvc *service.VersionService) *VersionHandler {
	return &VersionHandler{
		versionSvc: versionSvc,
	}
}

// List 获取版本列表
// GET /api/configs/:id/versions
func (h *VersionHandler) List(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	versions, err := h.versionSvc.List(c.Request.Context(), configID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"versions": versions,
	})
}

// Get 获取指定版本
// GET /api/configs/:id/versions/:version
func (h *VersionHandler) Get(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	versionNum, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的版本号",
		})
		return
	}

	version, err := h.versionSvc.GetByVersion(c.Request.Context(), configID, versionNum)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"version": version,
	})
}


// Diff 版本对比
// GET /api/configs/:id/diff?from=1&to=2
func (h *VersionHandler) Diff(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	fromV, err := strconv.Atoi(c.Query("from"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的起始版本号",
		})
		return
	}

	toV, err := strconv.Atoi(c.Query("to"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的目标版本号",
		})
		return
	}

	diff, err := h.versionSvc.Diff(c.Request.Context(), configID, fromV, toV)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"diff": diff,
	})
}

// Rollback 回滚到指定版本
// POST /api/configs/:id/rollback/:version
func (h *VersionHandler) Rollback(c *gin.Context) {
	configID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的配置 ID",
		})
		return
	}

	toVersion, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "无效的版本号",
		})
		return
	}

	userID := getUserID(c)
	author := "user"
	if userID > 0 {
		author = strconv.FormatInt(userID, 10)
	}

	version, err := h.versionSvc.Rollback(c.Request.Context(), configID, toVersion, author)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"version": version,
		"message": "回滚成功",
	})
}
