package repository

import (
	"context"

	"confighub/internal/model"

	"gorm.io/gorm"
)

// ConfigRepository 配置数据访问
type ConfigRepository struct {
	db *gorm.DB
}

// NewConfigRepository 创建配置仓库
func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// Create 创建配置
func (r *ConfigRepository) Create(ctx context.Context, config *model.Config) error {
	return r.db.WithContext(ctx).Create(config).Error
}

// GetByID 根据 ID 获取配置
func (r *ConfigRepository) GetByID(ctx context.Context, id int64) (*model.Config, error) {
	var config model.Config
	err := r.db.WithContext(ctx).First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetByProjectAndName 根据项目和名称获取配置
func (r *ConfigRepository) GetByProjectAndName(ctx context.Context, projectID int64, name string) (*model.Config, error) {
	var config model.Config
	err := r.db.WithContext(ctx).Where("project_id = ? AND name = ?", projectID, name).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetByProjectNamespaceEnv 根据项目、命名空间和环境获取配置
func (r *ConfigRepository) GetByProjectNamespaceEnv(ctx context.Context, projectID int64, namespace, env, name string) (*model.Config, error) {
	var config model.Config
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND namespace = ? AND environment = ? AND name = ?", projectID, namespace, env, name).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// List 获取项目下的配置列表
func (r *ConfigRepository) List(ctx context.Context, projectID int64) ([]*model.Config, error) {
	var configs []*model.Config
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("name ASC").Find(&configs).Error
	return configs, err
}

// ListByNamespace 根据命名空间获取配置列表
func (r *ConfigRepository) ListByNamespace(ctx context.Context, projectID int64, namespace string) ([]*model.Config, error) {
	var configs []*model.Config
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND namespace = ?", projectID, namespace).
		Order("environment ASC, name ASC").
		Find(&configs).Error
	return configs, err
}

// Update 更新配置
func (r *ConfigRepository) Update(ctx context.Context, config *model.Config) error {
	return r.db.WithContext(ctx).Save(config).Error
}

// Delete 删除配置
func (r *ConfigRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Config{}, id).Error
}

// IncrementVersion 增加版本号
func (r *ConfigRepository) IncrementVersion(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.Config{}).Where("id = ?", id).
		UpdateColumn("current_version", gorm.Expr("current_version + 1")).Error
}

// GetByNameAndEnv 根据名称和环境获取配置
func (r *ConfigRepository) GetByNameAndEnv(ctx context.Context, projectID int64, name, namespace, env string) (*model.Config, error) {
	var config model.Config
	query := r.db.WithContext(ctx).Where("project_id = ? AND name = ?", projectID, name)
	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}
	if env != "" {
		query = query.Where("environment = ?", env)
	} else {
		query = query.Where("environment = '' OR environment IS NULL")
	}
	err := query.First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}
