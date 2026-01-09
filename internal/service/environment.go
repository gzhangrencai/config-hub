package service

import (
	"context"
	"encoding/json"
	"errors"

	"confighub/internal/model"
	"confighub/internal/repository"
)

var (
	ErrEnvironmentNotFound = errors.New("环境不存在")
	ErrEnvironmentExists   = errors.New("环境已存在")
)

// Environment 环境定义
type Environment struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	IsDefault   bool   `json:"is_default"`
}

// DefaultEnvironments 默认环境列表
var DefaultEnvironments = []Environment{
	{Name: "dev", Description: "开发环境", Order: 1, IsDefault: true},
	{Name: "test", Description: "测试环境", Order: 2, IsDefault: true},
	{Name: "staging", Description: "预发布环境", Order: 3, IsDefault: true},
	{Name: "prod", Description: "生产环境", Order: 4, IsDefault: true},
}

// EnvironmentService 环境服务
type EnvironmentService struct {
	projectRepo *repository.ProjectRepository
	configRepo  *repository.ConfigRepository
	versionRepo *repository.VersionRepository
}

// NewEnvironmentService 创建环境服务
func NewEnvironmentService(projectRepo *repository.ProjectRepository, configRepo *repository.ConfigRepository, versionRepo *repository.VersionRepository) *EnvironmentService {
	return &EnvironmentService{
		projectRepo: projectRepo,
		configRepo:  configRepo,
		versionRepo: versionRepo,
	}
}

// List 获取项目环境列表
func (s *EnvironmentService) List(ctx context.Context, projectID int64) ([]Environment, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	// 如果项目有自定义环境配置
	if project.Settings != "" {
		var settings struct {
			Environments []Environment `json:"environments"`
		}
		if err := json.Unmarshal([]byte(project.Settings), &settings); err == nil && len(settings.Environments) > 0 {
			return settings.Environments, nil
		}
	}

	return DefaultEnvironments, nil
}

// Create 创建自定义环境
func (s *EnvironmentService) Create(ctx context.Context, projectID int64, env Environment) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return ErrProjectNotFound
	}

	envs, _ := s.List(ctx, projectID)
	for _, e := range envs {
		if e.Name == env.Name {
			return ErrEnvironmentExists
		}
	}

	env.Order = len(envs) + 1
	envs = append(envs, env)

	settings := map[string]interface{}{
		"environments": envs,
	}
	settingsJSON, _ := json.Marshal(settings)
	project.Settings = string(settingsJSON)

	return s.projectRepo.Update(ctx, project)
}

// MergeConfig 合并配置 (基础配置 + 环境覆盖)
func (s *EnvironmentService) MergeConfig(ctx context.Context, baseContent, envContent string) (string, error) {
	var base, env map[string]interface{}

	if err := json.Unmarshal([]byte(baseContent), &base); err != nil {
		return "", errors.New("基础配置格式无效")
	}

	if envContent == "" {
		return baseContent, nil
	}

	if err := json.Unmarshal([]byte(envContent), &env); err != nil {
		return "", errors.New("环境配置格式无效")
	}

	merged := s.deepMerge(base, env)
	result, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// deepMerge 深度合并两个 map
func (s *EnvironmentService) deepMerge(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 复制 base
	for k, v := range base {
		result[k] = v
	}

	// 合并 override
	for k, v := range override {
		if baseVal, exists := result[k]; exists {
			baseMap, baseIsMap := baseVal.(map[string]interface{})
			overrideMap, overrideIsMap := v.(map[string]interface{})
			if baseIsMap && overrideIsMap {
				result[k] = s.deepMerge(baseMap, overrideMap)
				continue
			}
		}
		result[k] = v
	}

	return result
}

// GetConfigForEnv 获取指定环境的配置
func (s *EnvironmentService) GetConfigForEnv(ctx context.Context, configID int64, env string) (*model.ConfigVersion, error) {
	config, err := s.configRepo.GetByID(ctx, configID)
	if err != nil {
		return nil, ErrConfigNotFound
	}

	// 获取当前版本
	version, err := s.versionRepo.GetLatest(ctx, configID)
	if err != nil {
		return nil, err
	}

	// 如果配置环境匹配，直接返回
	if config.Environment == env || config.Environment == "" {
		return version, nil
	}

	// 查找基础配置
	baseConfig, err := s.configRepo.GetByNameAndEnv(ctx, config.ProjectID, config.Name, config.Namespace, "")
	if err != nil {
		return version, nil
	}

	baseVersion, err := s.versionRepo.GetLatest(ctx, baseConfig.ID)
	if err != nil {
		return version, nil
	}

	// 合并配置
	mergedContent, err := s.MergeConfig(ctx, baseVersion.Content, version.Content)
	if err != nil {
		return version, nil
	}

	version.Content = mergedContent
	return version, nil
}

