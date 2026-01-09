package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"confighub/internal/model"
	"confighub/internal/repository"

	"gopkg.in/yaml.v3"
)

var (
	ErrConfigNotFound     = errors.New("配置不存在")
	ErrConfigNameExists   = errors.New("配置名称已存在")
	ErrInvalidJSON        = errors.New("无效的 JSON 格式")
	ErrInvalidYAML        = errors.New("无效的 YAML 格式")
	ErrInvalidFileType    = errors.New("不支持的文件类型")
)

// ConfigService 配置服务
type ConfigService struct {
	configRepo  *repository.ConfigRepository
	versionRepo *repository.VersionRepository
	projectRepo *repository.ProjectRepository
}

// NewConfigService 创建配置服务
func NewConfigService(configRepo *repository.ConfigRepository, versionRepo *repository.VersionRepository, projectRepo *repository.ProjectRepository) *ConfigService {
	return &ConfigService{
		configRepo:  configRepo,
		versionRepo: versionRepo,
		projectRepo: projectRepo,
	}
}

// UploadRequest 上传配置请求
type UploadRequest struct {
	Name        string `json:"name" binding:"required"`
	Namespace   string `json:"namespace"`
	Environment string `json:"environment"`
	FileType    string `json:"file_type" binding:"required"`
	Content     string `json:"content" binding:"required"`
	Message     string `json:"message"`
}

// Upload 上传配置
func (s *ConfigService) Upload(ctx context.Context, projectID int64, req *UploadRequest, author string) (*model.Config, error) {
	// 验证文件类型
	if req.FileType != "json" && req.FileType != "yaml" && req.FileType != "protobuf" {
		return nil, ErrInvalidFileType
	}

	// 解析和验证内容
	content := req.Content
	if req.FileType == "json" {
		if !json.Valid([]byte(content)) {
			return nil, ErrInvalidJSON
		}
	} else if req.FileType == "yaml" {
		// YAML 转 JSON
		var data interface{}
		if err := yaml.Unmarshal([]byte(content), &data); err != nil {
			return nil, ErrInvalidYAML
		}
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, ErrInvalidYAML
		}
		content = string(jsonBytes)
	}

	// 默认值
	namespace := req.Namespace
	if namespace == "" {
		namespace = "application"
	}
	environment := req.Environment
	if environment == "" {
		environment = "default"
	}

	// 检查是否已存在
	existing, _ := s.configRepo.GetByProjectNamespaceEnv(ctx, projectID, namespace, environment, req.Name)
	if existing != nil {
		return nil, ErrConfigNameExists
	}

	// 创建配置
	config := &model.Config{
		ProjectID:       projectID,
		Name:            req.Name,
		Namespace:       namespace,
		Environment:     environment,
		FileType:        req.FileType,
		DefaultEditMode: "code",
		CurrentVersion:  1,
	}

	if err := s.configRepo.Create(ctx, config); err != nil {
		return nil, err
	}

	// 创建初始版本
	commitHash := generateHash(content)
	message := req.Message
	if message == "" {
		message = "初始版本"
	}

	version := &model.ConfigVersion{
		ConfigID:      config.ID,
		Version:       1,
		Content:       content,
		CommitHash:    commitHash,
		CommitMessage: message,
		Author:        author,
	}

	if err := s.versionRepo.Create(ctx, version); err != nil {
		return nil, err
	}

	return config, nil
}

// GetByID 根据 ID 获取配置及最新版本
func (s *ConfigService) GetByID(ctx context.Context, id int64) (*model.Config, *model.ConfigVersion, error) {
	config, err := s.configRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, ErrConfigNotFound
	}

	version, err := s.versionRepo.GetLatest(ctx, id)
	if err != nil {
		return config, nil, nil
	}

	return config, version, nil
}

// GetConfigByID 仅获取配置信息
func (s *ConfigService) GetConfigByID(ctx context.Context, id int64) (*model.Config, error) {
	return s.configRepo.GetByID(ctx, id)
}

// List 获取项目下的配置列表
func (s *ConfigService) List(ctx context.Context, projectID int64) ([]*model.Config, error) {
	return s.configRepo.List(ctx, projectID)
}

// Update 更新配置内容
func (s *ConfigService) Update(ctx context.Context, id int64, content, message, author string) (*model.ConfigVersion, error) {
	config, err := s.configRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrConfigNotFound
	}

	// 验证 JSON
	if config.FileType == "json" && !json.Valid([]byte(content)) {
		return nil, ErrInvalidJSON
	}

	// 增加版本号
	newVersion := config.CurrentVersion + 1
	commitHash := generateHash(content)

	if message == "" {
		message = "更新配置"
	}

	version := &model.ConfigVersion{
		ConfigID:      id,
		Version:       newVersion,
		Content:       content,
		CommitHash:    commitHash,
		CommitMessage: message,
		Author:        author,
	}

	if err := s.versionRepo.Create(ctx, version); err != nil {
		return nil, err
	}

	// 更新配置的当前版本
	config.CurrentVersion = newVersion
	if err := s.configRepo.Update(ctx, config); err != nil {
		return nil, err
	}

	return version, nil
}

// Delete 删除配置
func (s *ConfigService) Delete(ctx context.Context, id int64) error {
	_, err := s.configRepo.GetByID(ctx, id)
	if err != nil {
		return ErrConfigNotFound
	}
	return s.configRepo.Delete(ctx, id)
}

// GetByAccessKey 通过 Access Key 获取配置
func (s *ConfigService) GetByAccessKey(ctx context.Context, projectID int64, configName, namespace, env string) (*model.Config, *model.ConfigVersion, error) {
	if namespace == "" {
		namespace = "application"
	}
	if env == "" {
		env = "default"
	}

	config, err := s.configRepo.GetByProjectNamespaceEnv(ctx, projectID, namespace, env, configName)
	if err != nil {
		return nil, nil, ErrConfigNotFound
	}

	version, err := s.versionRepo.GetLatest(ctx, config.ID)
	if err != nil {
		return config, nil, nil
	}

	return config, version, nil
}

// generateHash 生成内容哈希
func generateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])[:16]
}