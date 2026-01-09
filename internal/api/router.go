package api

import (
	"confighub/internal/config"
	"confighub/internal/middleware"
	"confighub/internal/repository"
	"confighub/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(router *gin.Engine, db *gorm.DB, rdb *redis.Client, logger *zap.Logger, cfg *config.Config) {
	// 初始化 Repository
	projectRepo := repository.NewProjectRepository(db)
	configRepo := repository.NewConfigRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	keyRepo := repository.NewKeyRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	releaseRepo := repository.NewReleaseRepository(db)

	// 初始化 Service
	projectSvc := service.NewProjectService(projectRepo, keyRepo)
	configSvc := service.NewConfigService(configRepo, versionRepo, projectRepo)
	versionSvc := service.NewVersionService(versionRepo, configRepo)
	schemaSvc := service.NewSchemaService(configRepo, versionRepo)
	keySvc := service.NewKeyService(keyRepo)
	auditSvc := service.NewAuditService(auditRepo)
	encryptSvc := service.NewEncryptionService(cfg.Encrypt.Key)
	releaseSvc := service.NewReleaseService(releaseRepo, configRepo, versionRepo)
	notifySvc := service.NewNotificationService(rdb)
	grayReleaseSvc := service.NewGrayReleaseService(releaseRepo, configRepo, versionRepo)
	envSvc := service.NewEnvironmentService(projectRepo, configRepo, versionRepo)
	envDiffSvc := service.NewEnvDiffService(configRepo, versionRepo)

	// 初始化 Handler
	projectHandler := NewProjectHandler(projectSvc, auditSvc)
	configHandler := NewConfigHandler(configSvc, auditSvc)
	versionHandler := NewVersionHandler(versionSvc)
	schemaHandler := NewSchemaHandler(schemaSvc)
	keyHandler := NewKeyHandler(keySvc, auditSvc)
	auditHandler := NewAuditHandler(auditSvc)
	releaseHandler := NewReleaseHandler(releaseSvc, grayReleaseSvc, auditSvc)
	publicConfigHandler := NewPublicConfigHandler(configSvc, encryptSvc, notifySvc, auditSvc)
	envHandler := NewEnvironmentHandler(envSvc, envDiffSvc)

	// 根路径 - API 信息
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "ConfigHub",
			"version": "1.0.0",
			"status":  "running",
			"endpoints": gin.H{
				"health":   "/health",
				"api":      "/api/v1",
				"docs":     "https://github.com/gzhangrencai/config-hub",
			},
		})
	})

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "confighub"})
	})

	// API v1 - 公开配置接口 (客户端使用)
	v1 := router.Group("/api/v1")
	{
		v1.Use(middleware.OptionalAuth(db, cfg.JWT.Secret))
		v1.GET("/config", publicConfigHandler.Get)
		v1.PUT("/config", middleware.RequirePermission("write"), publicConfigHandler.Update)
		v1.POST("/config", middleware.RequirePermission("write"), publicConfigHandler.Create)
		v1.GET("/config/watch", publicConfigHandler.Watch)
	}

	// API - 管理接口
	api := router.Group("/api")
	{
		// 项目管理
		projects := api.Group("/projects")
		projects.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			projects.POST("", projectHandler.Create)
			projects.GET("", projectHandler.List)
			projects.GET("/:id", projectHandler.Get)
			projects.PUT("/:id", projectHandler.Update)
			projects.DELETE("/:id", projectHandler.Delete)

			// 项目下的配置
			projects.POST("/:id/configs", configHandler.Upload)
			projects.GET("/:id/configs", configHandler.List)

			// 项目下的密钥
			projects.POST("/:id/keys", keyHandler.Create)
			projects.GET("/:id/keys", keyHandler.List)

			// 项目下的审计日志
			projects.GET("/:id/audit-logs", auditHandler.List)

			// 项目下的环境
			projects.GET("/:id/environments", envHandler.List)
			projects.POST("/:id/environments", envHandler.Create)
		}

		// 配置管理
		configs := api.Group("/configs")
		configs.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			configs.GET("/:id", configHandler.Get)
			configs.PUT("/:id", configHandler.Update)
			configs.DELETE("/:id", configHandler.Delete)

			// 版本管理
			configs.GET("/:id/versions", versionHandler.List)
			configs.GET("/:id/versions/:version", versionHandler.Get)
			configs.GET("/:id/diff", versionHandler.Diff)
			configs.POST("/:id/rollback/:version", versionHandler.Rollback)

			// Schema 管理
			configs.GET("/:id/schema", schemaHandler.Get)
			configs.PUT("/:id/schema", schemaHandler.Update)
			configs.POST("/:id/schema/generate", schemaHandler.Generate)

			// 发布管理
			configs.POST("/:id/release", releaseHandler.Create)
			configs.GET("/:id/releases", releaseHandler.List)
			configs.POST("/:id/gray-release", releaseHandler.CreateGray)

			// 环境对比
			configs.GET("/:id/compare", envHandler.Compare)
			configs.POST("/:id/sync", envHandler.Sync)
			configs.POST("/:id/merge", envHandler.MergeConfig)
		}

		// 密钥管理
		keys := api.Group("/keys")
		keys.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			keys.PUT("/:id", keyHandler.Update)
			keys.DELETE("/:id", keyHandler.Delete)
			keys.POST("/:id/regenerate", keyHandler.Regenerate)
		}

		// 发布管理
		releases := api.Group("/releases")
		releases.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			releases.POST("/:id/rollback", releaseHandler.Rollback)
			releases.POST("/:id/promote", releaseHandler.Promote)
			releases.POST("/:id/cancel", releaseHandler.Cancel)
			releases.PUT("/:id/percentage", releaseHandler.UpdateGrayPercentage)
		}

		// 用户认证
		auth := api.Group("/auth")
		{
			auth.POST("/login", projectHandler.Login)
			auth.POST("/register", projectHandler.Register)
		}
	}
}
