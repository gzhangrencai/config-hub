package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"confighub/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	// SignatureHeader 签名头
	SignatureHeader = "X-Signature"
	// TimestampHeader 时间戳头
	TimestampHeader = "X-Timestamp"
	// NonceHeader 随机数头
	NonceHeader = "X-Nonce"
	// MaxTimeDiff 最大时间差 (5分钟)
	MaxTimeDiff = 5 * 60
)

// SignatureAuth 签名认证中间件
func SignatureAuth(db *gorm.DB) gin.HandlerFunc {
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

		signature := c.GetHeader(SignatureHeader)
		timestamp := c.GetHeader(TimestampHeader)
		nonce := c.GetHeader(NonceHeader)

		// 如果没有签名头，跳过签名验证 (向后兼容)
		if signature == "" {
			c.Next()
			return
		}

		// 验证时间戳
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_SIGNATURE",
				"message": "无效的时间戳",
			})
			return
		}

		now := time.Now().Unix()
		if abs(now-ts) > MaxTimeDiff {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_SIGNATURE",
				"message": "请求已过期",
			})
			return
		}

		// 获取密钥
		var key model.ProjectKey
		if err := db.Where("access_key = ? AND is_active = ?", accessKey, true).First(&key).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "无效的 Access Key",
			})
			return
		}

		// 构建签名字符串
		stringToSign := buildStringToSign(c, accessKey, timestamp, nonce)

		// 验证签名
		expectedSignature := calculateSignature(stringToSign, key.SecretKeyHash)
		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "INVALID_SIGNATURE",
				"message": "签名验证失败",
			})
			return
		}

		c.Next()
	}
}

// buildStringToSign 构建待签名字符串
func buildStringToSign(c *gin.Context, accessKey, timestamp, nonce string) string {
	var parts []string

	// HTTP 方法
	parts = append(parts, c.Request.Method)

	// 请求路径
	parts = append(parts, c.Request.URL.Path)

	// 查询参数 (按字母排序)
	queryParams := c.Request.URL.Query()
	var sortedKeys []string
	for k := range queryParams {
		if k != "signature" {
			sortedKeys = append(sortedKeys, k)
		}
	}
	sort.Strings(sortedKeys)

	var queryParts []string
	for _, k := range sortedKeys {
		for _, v := range queryParams[k] {
			queryParts = append(queryParts, k+"="+v)
		}
	}
	parts = append(parts, strings.Join(queryParts, "&"))

	// Access Key
	parts = append(parts, accessKey)

	// 时间戳
	parts = append(parts, timestamp)

	// 随机数
	parts = append(parts, nonce)

	return strings.Join(parts, "\n")
}

// calculateSignature 计算签名
func calculateSignature(stringToSign, secretKey string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(stringToSign))
	return hex.EncodeToString(h.Sum(nil))
}

// abs 绝对值
func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
