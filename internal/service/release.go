package service

import (
	"context"
	"errors"

	"confighub/internal/model"
	"confighub/internal/repository"
)

var (
	ErrReleaseNotFound = errors.New("发布记录不存在")
)

// ReleaseService 发布服务
type ReleaseService struct {
	releaseRepo *repository.ReleaseRepository
	configRepo  *repository.ConfigRepository
	versionRepo *repository.VersionRepository
}

// NewReleaseService 创建发布服务
func NewReleaseService(releaseRepo *repository.ReleaseRepository, configRepo *repository.ConfigRepository, versionRepo *repository.VersionRepository) *ReleaseService {
	return &ReleaseService{
		releaseRepo: releaseRepo,
		configRepo:  configRepo,
		versionRepo: versionRepo,
	}
}

// Create 创建发布
func (s *ReleaseService) Create(ctx context.Context, configID int64, env string, version int, author string) (*model.Release, error) {
	config, err := s.configRepo.GetByID(ctx, configID)
	if err != nil {
		return nil, ErrConfigNotFound
	}

	// 如果未指定版本，使用当前版本
	if version == 0 {
		version = config.CurrentVersion
	}

	// 验证版本存在
	_, err = s.versionRepo.GetByConfigAndVersion(ctx, configID, version)
	if err != nil {
		return nil, ErrVersionNotFound
	}

	release := &model.Release{
		ProjectID:   config.ProjectID,
		ConfigID:    configID,
		Version:     version,
		Environment: env,
		Status:      "released",
		ReleaseType: "full",
		ReleasedBy:  author,
	}

	if err := s.releaseRepo.Create(ctx, release); err != nil {
		return nil, err
	}

	return release, nil
}

// List 获取发布历史
func (s *ReleaseService) List(ctx context.Context, configID int64) ([]*model.Release, error) {
	return s.releaseRepo.List(ctx, configID)
}

// Rollback 回滚发布
func (s *ReleaseService) Rollback(ctx context.Context, releaseID int64, author string) (*model.Release, error) {
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		return nil, ErrReleaseNotFound
	}

	// 标记当前发布为回滚状态
	release.Status = "rollback"
	if err := s.releaseRepo.Update(ctx, release); err != nil {
		return nil, err
	}

	// 获取上一个发布
	releases, err := s.releaseRepo.List(ctx, release.ConfigID)
	if err != nil || len(releases) < 2 {
		return nil, errors.New("没有可回滚的版本")
	}

	// 找到上一个发布的版本
	var prevVersion int
	for _, r := range releases {
		if r.ID != releaseID && r.Environment == release.Environment && r.Status == "released" {
			prevVersion = r.Version
			break
		}
	}

	if prevVersion == 0 {
		return nil, errors.New("没有可回滚的版本")
	}

	// 创建新的发布记录
	newRelease := &model.Release{
		ProjectID:   release.ProjectID,
		ConfigID:    release.ConfigID,
		Version:     prevVersion,
		Environment: release.Environment,
		Status:      "released",
		ReleaseType: "full",
		ReleasedBy:  author,
	}

	if err := s.releaseRepo.Create(ctx, newRelease); err != nil {
		return nil, err
	}

	return newRelease, nil
}

// GetByEnv 获取指定环境的当前发布
func (s *ReleaseService) GetByEnv(ctx context.Context, configID int64, env string) (*model.Release, error) {
	return s.releaseRepo.GetByConfigAndEnv(ctx, configID, env)
}

// ListEnvironments 获取项目环境列表
func (s *ReleaseService) ListEnvironments(ctx context.Context, projectID int64) ([]string, error) {
	// 返回默认环境列表
	return []string{"dev", "test", "staging", "prod"}, nil
}

// CreateEnvironment 创建环境
func (s *ReleaseService) CreateEnvironment(ctx context.Context, projectID int64, name, description string) error {
	// 简化实现，实际应该存储到数据库
	return nil
}
