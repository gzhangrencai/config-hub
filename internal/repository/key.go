package repository

import (
	"context"

	"confighub/internal/model"

	"gorm.io/gorm"
)

// KeyRepository 密钥数据访问
type KeyRepository struct {
	db *gorm.DB
}

// NewKeyRepository 创建密钥仓库
func NewKeyRepository(db *gorm.DB) *KeyRepository {
	return &KeyRepository{db: db}
}

// Create 创建密钥
func (r *KeyRepository) Create(ctx context.Context, key *model.ProjectKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

// GetByID 根据 ID 获取密钥
func (r *KeyRepository) GetByID(ctx context.Context, id int64) (*model.ProjectKey, error) {
	var key model.ProjectKey
	err := r.db.WithContext(ctx).First(&key, id).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// GetByAccessKey 根据 Access Key 获取密钥
func (r *KeyRepository) GetByAccessKey(ctx context.Context, accessKey string) (*model.ProjectKey, error) {
	var key model.ProjectKey
	err := r.db.WithContext(ctx).Where("access_key = ?", accessKey).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// List 获取项目下的密钥列表
func (r *KeyRepository) List(ctx context.Context, projectID int64) ([]*model.ProjectKey, error) {
	var keys []*model.ProjectKey
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").Find(&keys).Error
	return keys, err
}

// Update 更新密钥
func (r *KeyRepository) Update(ctx context.Context, key *model.ProjectKey) error {
	return r.db.WithContext(ctx).Save(key).Error
}

// Delete 删除密钥
func (r *KeyRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.ProjectKey{}, id).Error
}

// Deactivate 停用密钥
func (r *KeyRepository) Deactivate(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&model.ProjectKey{}).Where("id = ?", id).Update("is_active", false).Error
}
