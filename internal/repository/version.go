package repository

import (
	"context"

	"confighub/internal/model"

	"gorm.io/gorm"
)

// VersionRepository 版本数据访问
type VersionRepository struct {
	db *gorm.DB
}

// NewVersionRepository 创建版本仓库
func NewVersionRepository(db *gorm.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

// Create 创建版本
func (r *VersionRepository) Create(ctx context.Context, version *model.ConfigVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

// GetByConfigAndVersion 根据配置 ID 和版本号获取版本
func (r *VersionRepository) GetByConfigAndVersion(ctx context.Context, configID int64, version int) (*model.ConfigVersion, error) {
	var v model.ConfigVersion
	err := r.db.WithContext(ctx).Where("config_id = ? AND version = ?", configID, version).First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// List 获取配置的所有版本
func (r *VersionRepository) List(ctx context.Context, configID int64) ([]*model.ConfigVersion, error) {
	var versions []*model.ConfigVersion
	err := r.db.WithContext(ctx).Where("config_id = ?", configID).Order("version DESC").Find(&versions).Error
	return versions, err
}

// GetLatest 获取最新版本
func (r *VersionRepository) GetLatest(ctx context.Context, configID int64) (*model.ConfigVersion, error) {
	var version model.ConfigVersion
	err := r.db.WithContext(ctx).Where("config_id = ?", configID).Order("version DESC").First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

// GetVersionsSince 获取指定版本之后的所有版本
func (r *VersionRepository) GetVersionsSince(ctx context.Context, configID int64, sinceVersion int) ([]*model.ConfigVersion, error) {
	var versions []*model.ConfigVersion
	err := r.db.WithContext(ctx).
		Where("config_id = ? AND version > ?", configID, sinceVersion).
		Order("version ASC").
		Find(&versions).Error
	return versions, err
}

// DeleteByConfigID 删除配置的所有版本
func (r *VersionRepository) DeleteByConfigID(ctx context.Context, configID int64) error {
	return r.db.WithContext(ctx).Where("config_id = ?", configID).Delete(&model.ConfigVersion{}).Error
}

// Update 更新版本
func (r *VersionRepository) Update(ctx context.Context, version *model.ConfigVersion) error {
	return r.db.WithContext(ctx).Save(version).Error
}
