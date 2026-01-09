package repository

import (
	"context"

	"confighub/internal/model"

	"gorm.io/gorm"
)

// ProjectRepository 项目数据访问
type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository 创建项目仓库
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Create 创建项目
func (r *ProjectRepository) Create(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

// GetByID 根据 ID 获取项目
func (r *ProjectRepository) GetByID(ctx context.Context, id int64) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetByName 根据名称获取项目
func (r *ProjectRepository) GetByName(ctx context.Context, name string) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// List 获取项目列表
func (r *ProjectRepository) List(ctx context.Context, userID int64) ([]*model.Project, error) {
	var projects []*model.Project
	query := r.db.WithContext(ctx)
	if userID > 0 {
		query = query.Where("created_by = ?", userID)
	}
	err := query.Order("created_at DESC").Find(&projects).Error
	return projects, err
}

// Update 更新项目
func (r *ProjectRepository) Update(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

// Delete 删除项目
func (r *ProjectRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Project{}, id).Error
}

// ExistsByName 检查项目名是否存在
func (r *ProjectRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Project{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// CreateEnvironment 创建环境
func (r *ProjectRepository) CreateEnvironment(ctx context.Context, env *model.ProjectEnvironment) error {
	return r.db.WithContext(ctx).Create(env).Error
}

// ListEnvironments 获取项目环境列表
func (r *ProjectRepository) ListEnvironments(ctx context.Context, projectID int64) ([]*model.ProjectEnvironment, error) {
	var envs []*model.ProjectEnvironment
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("sort_order ASC").Find(&envs).Error
	return envs, err
}
