package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"confighub/internal/middleware"
	"confighub/internal/model"
	"confighub/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PublicConfigHandler 公开配置 API 处理器
type PublicConfigHandler struct {
	configSvc  *service.ConfigService
	encryptSvc *service.EncryptionService
	notifySvc  *service.NotificationService
	auditSvc   *service.AuditService
}

// NewPublicConfigHandler 创建公开配置处理器
func NewPublicConfigHandler(configSvc *service.ConfigService, encryptSvc *service.EncryptionService, notifySvc *service.NotificationService, auditSvc *service.AuditService) *PublicConfigHandler {
	return &PublicConfigHandler{
		configSvc:  configSvc,
		encryptSvc: encryptSvc,
		notifySvc:  notifySvc,
		auditSvc:   auditSvc,
	}
}

// Get 获取配置
// GET /api/v1/config?name=xxx&namespace=xxx&env=xxx
func (h *PublicConfigHandler) Get(c *gin.Context) {
	projectID := getProjectID(c)
	if projectID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "未授权访问",
		})
		return
	}

	configName := c.Query("name")
	if configName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "配置名称不能为空",
		})
		return
	}

	namespace := c.Query("namespace")
	env := c.Query("env")

	config, version, err := h.configSvc.GetByAccessKey(c.Request.Context(), projectID, configName, namespace, env)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "配置不存在",
		})
		return
	}

	response := gin.H{
		"name":        config.Name,
		"namespace":   config.Namespace,
		"environment": config.Environment,
		"version":     config.CurrentVersion,
	}

	if version != nil {
		content := version.Content
		authCtx := middleware.GetAuthContext(c)
		if authCtx != nil && authCtx.Permissions.Decrypt {
			content = h.decryptSensitiveFields(content)
		}
		response["content"] = content
	}

	h.logAccess(c, projectID, config.ID, "read")
	c.JSON(http.StatusOK, response)
}


// Update 更新配置
// PUT /api/v1/config
func (h *PublicConfigHandler) Update(c *gin.Context) {
	projectID := getProjectID(c)
	if projectID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "未授权访问",
		})
		return
	}

	var req struct {
		Name      string `json:"name" binding:"required"`
		Namespace string `json:"namespace"`
		Env       string `json:"env"`
		Content   string `json:"content" binding:"required"`
		Message   string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	config, _, err := h.configSvc.GetByAccessKey(c.Request.Context(), projectID, req.Name, req.Namespace, req.Env)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "配置不存在",
		})
		return
	}

	author := "api"
	authCtx := middleware.GetAuthContext(c)
	if authCtx != nil && authCtx.AccessKeyID > 0 {
		author = "key:" + strconv.FormatInt(authCtx.AccessKeyID, 10)
	}

	message := req.Message
	if message == "" {
		message = "通过 API 更新"
	}

	version, err := h.configSvc.Update(c.Request.Context(), config.ID, req.Content, message, author)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	h.logAccess(c, projectID, config.ID, "update")

	h.notifySvc.NotifyChange(c.Request.Context(), &service.ConfigChange{
		ConfigID:   config.ID,
		ConfigName: config.Name,
		Namespace:  config.Namespace,
		Env:        config.Environment,
		Version:    version.Version,
		ChangeType: "update",
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"version": version.Version,
	})
}

// Create 创建配置
// POST /api/v1/config
func (h *PublicConfigHandler) Create(c *gin.Context) {
	projectID := getProjectID(c)
	if projectID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "未授权访问",
		})
		return
	}

	var req service.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "请求参数无效",
			"details": err.Error(),
		})
		return
	}

	if req.FileType == "" {
		req.FileType = "json"
	}

	author := "api"
	authCtx := middleware.GetAuthContext(c)
	if authCtx != nil && authCtx.AccessKeyID > 0 {
		author = "key:" + strconv.FormatInt(authCtx.AccessKeyID, 10)
	}

	config, err := h.configSvc.Upload(c.Request.Context(), projectID, &req, author)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	h.logAccess(c, projectID, config.ID, "create")

	c.JSON(http.StatusCreated, gin.H{
		"message": "创建成功",
		"config": gin.H{
			"id":          config.ID,
			"name":        config.Name,
			"namespace":   config.Namespace,
			"environment": config.Environment,
			"version":     config.CurrentVersion,
		},
	})
}


// Watch 监听配置变更 (Long-Polling)
// GET /api/v1/config/watch?name=xxx&namespace=xxx&env=xxx&version=xxx&timeout=xxx
func (h *PublicConfigHandler) Watch(c *gin.Context) {
	projectID := getProjectID(c)
	if projectID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "未授权访问",
		})
		return
	}

	configName := c.Query("name")
	if configName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST",
			"message": "配置名称不能为空",
		})
		return
	}

	namespace := c.Query("namespace")
	env := c.Query("env")

	currentVersion := 0
	if v := c.Query("version"); v != "" {
		currentVersion, _ = strconv.Atoi(v)
	}

	timeout := 30
	if t := c.Query("timeout"); t != "" {
		timeout, _ = strconv.Atoi(t)
		if timeout > 60 {
			timeout = 60
		}
		if timeout < 1 {
			timeout = 1
		}
	}

	config, version, err := h.configSvc.GetByAccessKey(c.Request.Context(), projectID, configName, namespace, env)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    "NOT_FOUND",
			"message": "配置不存在",
		})
		return
	}

	if version != nil && version.Version > currentVersion {
		c.JSON(http.StatusOK, gin.H{
			"changed":     true,
			"name":        config.Name,
			"namespace":   config.Namespace,
			"environment": config.Environment,
			"version":     version.Version,
			"content":     version.Content,
		})
		return
	}

	clientID := uuid.New().String()
	changeCh, err := h.notifySvc.Subscribe(c.Request.Context(), clientID, []int64{config.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "订阅失败",
		})
		return
	}
	defer h.notifySvc.Unsubscribe(c.Request.Context(), clientID)

	select {
	case change := <-changeCh:
		if change != nil && change.ConfigID == config.ID {
			_, newVersion, _ := h.configSvc.GetByAccessKey(c.Request.Context(), projectID, configName, namespace, env)
			if newVersion != nil {
				c.JSON(http.StatusOK, gin.H{
					"changed":     true,
					"name":        config.Name,
					"namespace":   config.Namespace,
					"environment": config.Environment,
					"version":     newVersion.Version,
					"content":     newVersion.Content,
				})
				return
			}
		}
	case <-time.After(time.Duration(timeout) * time.Second):
		c.Status(http.StatusNotModified)
		return
	case <-c.Request.Context().Done():
		return
	}

	c.Status(http.StatusNotModified)
}

// decryptSensitiveFields 解密敏感字段
func (h *PublicConfigHandler) decryptSensitiveFields(content string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return content
	}

	h.decryptMap(data)

	result, err := json.Marshal(data)
	if err != nil {
		return content
	}
	return string(result)
}

// decryptMap 递归解密 map 中的加密字段
func (h *PublicConfigHandler) decryptMap(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			if len(v) > 4 && v[:4] == "ENC:" {
				decrypted, err := h.encryptSvc.Decrypt(v[4:])
				if err == nil {
					data[key] = decrypted
				}
			}
		case map[string]interface{}:
			h.decryptMap(v)
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					h.decryptMap(m)
				}
			}
		}
	}
}

// logAccess 记录访问日志
func (h *PublicConfigHandler) logAccess(c *gin.Context, projectID, configID int64, action string) {
	authCtx := middleware.GetAuthContext(c)
	var keyID *int64
	if authCtx != nil && authCtx.AccessKeyID > 0 {
		keyID = &authCtx.AccessKeyID
	}

	h.auditSvc.Log(c.Request.Context(), &model.AuditLog{
		ProjectID:    projectID,
		AccessKeyID:  keyID,
		Action:       action,
		ResourceType: model.AuditResourceConfig,
		ResourceID:   configID,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})
}
