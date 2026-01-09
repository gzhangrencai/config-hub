package middleware

import (
	"bytes"
	"io"
	"strings"

	"confighub/internal/model"
	"confighub/internal/service"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware 审计日志中间件
func AuditMiddleware(auditSvc *service.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查和静态资源
		if c.Request.URL.Path == "/health" || strings.HasPrefix(c.Request.URL.Path, "/static") {
			c.Next()
			return
		}

		// 跳过 GET 请求 (读取操作由具体 handler 记录)
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		// 读取请求体
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestBody = string(bodyBytes)
			// 重新设置请求体
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 处理请求
		c.Next()

		// 记录审计日志
		authCtx := GetAuthContext(c)
		if authCtx == nil {
			return
		}

		// 确定操作类型
		action := getActionFromMethod(c.Request.Method)
		resourceType, resourceID := parseResourceFromPath(c.Request.URL.Path)

		log := &model.AuditLog{
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			RequestBody:  truncateString(requestBody, 2000),
		}

		if authCtx.UserID > 0 {
			log.UserID = &authCtx.UserID
		}
		if authCtx.AccessKeyID > 0 {
			log.AccessKeyID = &authCtx.AccessKeyID
		}
		if authCtx.ProjectID > 0 {
			log.ProjectID = authCtx.ProjectID
		}

		auditSvc.Log(c.Request.Context(), log)
	}
}

// getActionFromMethod 根据 HTTP 方法获取操作类型
func getActionFromMethod(method string) string {
	switch method {
	case "POST":
		return model.AuditActionCreate
	case "PUT", "PATCH":
		return model.AuditActionUpdate
	case "DELETE":
		return model.AuditActionDelete
	default:
		return model.AuditActionRead
	}
}

// parseResourceFromPath 从路径解析资源类型和 ID
func parseResourceFromPath(path string) (string, int64) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	
	// 跳过 api 前缀
	if len(parts) > 0 && parts[0] == "api" {
		parts = parts[1:]
	}
	if len(parts) > 0 && (parts[0] == "v1" || parts[0] == "v2") {
		parts = parts[1:]
	}

	if len(parts) == 0 {
		return "", 0
	}

	resourceType := parts[0]
	var resourceID int64

	// 尝试解析资源 ID
	if len(parts) > 1 {
		// 简单解析，实际应该更精确
		for _, part := range parts[1:] {
			if id := parseID(part); id > 0 {
				resourceID = id
				break
			}
		}
	}

	// 标准化资源类型
	switch resourceType {
	case "projects":
		return model.AuditResourceProject, resourceID
	case "configs", "config":
		return model.AuditResourceConfig, resourceID
	case "keys":
		return model.AuditResourceKey, resourceID
	case "releases":
		return model.AuditResourceRelease, resourceID
	default:
		return resourceType, resourceID
	}
}

// parseID 解析 ID
func parseID(s string) int64 {
	var id int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			id = id*10 + int64(c-'0')
		} else {
			return 0
		}
	}
	return id
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
