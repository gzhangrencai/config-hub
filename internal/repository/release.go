package repository

import (
	"context"

	"confighub/internal/model"

	"gorm.io/gorm"
)

// ReleaseRepository 发布数据访问
type ReleaseRepository struct {
	db *gorm.DB
}

// NewReleaseRepository 创建发布仓库
func NewReleaseRepository(db *gorm.DB) *ReleaseRepository {
	return &ReleaseRepository{db: db}
}

// Create 创建发布记录
func (r *ReleaseRepository) Create(ctx context.Context, release *model.Release) error {
	return r.db.WithContext(ctx).Create(release).Error
}

// GetByID 根据 ID 获取发布记录
func (r *ReleaseRepository) GetByID(ctx context.Context, id int64) (*model.Release, error) {
	var release model.Release
	err := r.db.WithContext(ctx).First(&release, id).Error
	if err != nil {
		return nil, err
	}
	return &release, nil
}

// GetByConfigAndEnv 获取配置在指定环境的最新发布
func (r *ReleaseRepository) GetByConfigAndEnv(ctx context.Context, configID int64, env string) (*model.Release, error) {
	var release model.Release
	err := r.db.WithContext(ctx).
		Where("config_id = ? AND environment = ? AND status IN ('released', 'gray')", configID, env).
		Order("released_at DESC").
		First(&release).Error
	if err != nil {
		return nil, err
	}
	return &release, nil
}

// List 获取配置的发布历史
func (r *ReleaseRepository) List(ctx context.Context, configID int64) ([]*model.Release, error) {
	var releases []*model.Release
	err := r.db.WithContext(ctx).Where("config_id = ?", configID).Order("released_at DESC").Find(&releases).Error
	return releases, err
}

// ListByProject 获取项目的发布历史
func (r *ReleaseRepository) ListByProject(ctx context.Context, projectID int64) ([]*model.Release, error) {
	var releases []*model.Release
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("released_at DESC").Find(&releases).Error
	return releases, err
}

// Update 更新发布记录
func (r *ReleaseRepository) Update(ctx context.Context, release *model.Release) error {
	return r.db.WithContext(ctx).Save(release).Error
}

// GetActiveGrayRelease 获取活跃的灰度发布
func (r *ReleaseRepository) GetActiveGrayRelease(ctx context.Context, configID int64, env string) (*model.Release, error) {
	var release model.Release
	err := r.db.WithContext(ctx).
		Where("config_id = ? AND environment = ? AND status = 'gray'", configID, env).
		First(&release).Error
	if err != nil {
		return nil, err
	}
	return &release, nil
}
