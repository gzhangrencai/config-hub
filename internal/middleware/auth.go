package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"confighub/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// AuthContext 认证上下文
type AuthContext struct {
	UserID      int64
	Username    string
	AccessKeyID int64
	ProjectID   int64
	Permissions model.Permissions
}

const AuthContextKey = "auth_context"

// JWTAuth JWT 认证中间件
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "缺少认证信息",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "认证格式错误",
			})
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "无效的认证令牌",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "无效的认证令牌",
			})
			return
		}

		authCtx := &AuthContext{
			UserID:   int64(claims["user_id"].(float64)),
			Username: claims["username"].(string),
			Permissions: model.Permissions{
				Read:    true,
				Write:   true,
				Delete:  true,
				Release: true,
				Admin:   true,
				Decrypt: true,
			},
		}

		c.Set(AuthContextKey, authCtx)
		c.Next()
	}
}

// AccessKeyAuth Access Key 认证中间件
func AccessKeyAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessKey := c.GetHeader("X-Access-Key")
		if accessKey == "" {
			accessKey = c.Query("access_key")
		}

		if accessKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "缺少 Access Key",
			})
			return
		}

		var key model.ProjectKey
		if err := db.Where("access_key = ? AND is_active = ?", accessKey, true).First(&key).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "无效的 Access Key",
			})
			return
		}

		// 检查过期时间
		if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Access Key 已过期",
			})
			return
		}

		// 检查 IP 白名单
		if key.IPWhitelist != "" && key.IPWhitelist != "[]" {
			var whitelist []string
			if err := json.Unmarshal([]byte(key.IPWhitelist), &whitelist); err == nil && len(whitelist) > 0 {
				clientIP := c.ClientIP()
				allowed := false
				for _, ip := range whitelist {
					if ip == clientIP || matchIPRange(clientIP, ip) {
						allowed = true
						break
					}
				}
				if !allowed {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
						"code":    "FORBIDDEN",
						"message": "IP 地址不在白名单中",
					})
					return
				}
			}
		}

		// 解析权限
		var permissions model.Permissions
		if key.Permissions != "" {
			json.Unmarshal([]byte(key.Permissions), &permissions)
		}

		authCtx := &AuthContext{
			AccessKeyID: key.ID,
			ProjectID:   key.ProjectID,
			Permissions: permissions,
		}

		c.Set(AuthContextKey, authCtx)
		c.Next()
	}
}

// OptionalAuth 可选认证中间件 (用于公开模式)
func OptionalAuth(db *gorm.DB, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试 JWT 认证
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			JWTAuth(jwtSecret)(c)
			if c.IsAborted() {
				return
			}
			c.Next()
			return
		}

		// 尝试 Access Key 认证
		accessKey := c.GetHeader("X-Access-Key")
		if accessKey == "" {
			accessKey = c.Query("access_key")
		}
		if accessKey != "" {
			AccessKeyAuth(db)(c)
			if c.IsAborted() {
				return
			}
			c.Next()
			return
		}

		// 无认证，设置只读权限
		authCtx := &AuthContext{
			Permissions: model.Permissions{
				Read: true,
			},
		}
		c.Set(AuthContextKey, authCtx)
		c.Next()
	}
}

// RequirePermission 权限检查中间件
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, exists := c.Get(AuthContextKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "未认证",
			})
			return
		}

		ctx := authCtx.(*AuthContext)
		allowed := false

		switch permission {
		case "read":
			allowed = ctx.Permissions.Read
		case "write":
			allowed = ctx.Permissions.Write
		case "delete":
			allowed = ctx.Permissions.Delete
		case "release":
			allowed = ctx.Permissions.Release
		case "admin":
			allowed = ctx.Permissions.Admin
		case "decrypt":
			allowed = ctx.Permissions.Decrypt
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "FORBIDDEN",
				"message": "无权限执行此操作",
			})
			return
		}

		c.Next()
	}
}

// GetAuthContext 获取认证上下文
func GetAuthContext(c *gin.Context) *AuthContext {
	if authCtx, exists := c.Get(AuthContextKey); exists {
		return authCtx.(*AuthContext)
	}
	return nil
}

// matchIPRange IP 范围匹配 (支持 CIDR 和通配符)
func matchIPRange(clientIP, pattern string) bool {
	// 精确匹配
	if clientIP == pattern {
		return true
	}

	// CIDR 匹配
	if strings.Contains(pattern, "/") {
		_, ipNet, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		ip := net.ParseIP(clientIP)
		if ip == nil {
			return false
		}
		return ipNet.Contains(ip)
	}

	// 通配符匹配 (如 192.168.1.*)
	if strings.Contains(pattern, "*") {
		patternParts := strings.Split(pattern, ".")
		ipParts := strings.Split(clientIP, ".")
		if len(patternParts) != len(ipParts) {
			return false
		}
		for i, part := range patternParts {
			if part != "*" && part != ipParts[i] {
				return false
			}
		}
		return true
	}

	return false
}
