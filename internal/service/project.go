package service

import (
	"context"
	"encoding/json"
	"errors"

	"confighub/internal/model"
	"confighub/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrProjectNotFound      = errors.New("项目不存在")
	ErrProjectNameExists    = errors.New("项目名称已存在")
	ErrProjectNameRequired  = errors.New("项目名称不能为空")
)

// ProjectService 项目服务
type ProjectService struct {
	projectRepo *repository.ProjectRepository
	keyRepo     *repository.KeyRepository
}

// NewProjectService 创建项目服务
func NewProjectService(projectRepo *repository.ProjectRepository, keyRepo *repository.KeyRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		keyRepo:     keyRepo,
	}
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	AccessMode  string `json:"access_mode"`
	GitRepoURL  string `json:"git_repo_url"`
	GitBranch   string `json:"git_branch"`
}

// Create 创建项目
func (s *ProjectService) Create(ctx context.Context, req *CreateProjectRequest, userID int64) (*model.Project, *model.ProjectKey, error) {
	if req.Name == "" {
		return nil, nil, ErrProjectNameRequired
	}

	// 检查名称是否存在
	exists, err := s.projectRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, nil, err
	}
	if exists {
		return nil, nil, ErrProjectNameExists
	}

	// 默认访问模式
	accessMode := req.AccessMode
	if accessMode == "" {
		accessMode = "key"
	}

	// 默认权限
	publicPerms, _ := json.Marshal(map[string]bool{"read": true, "write": false})

	// 默认分支
	gitBranch := req.GitBranch
	if gitBranch == "" {
		gitBranch = "main"
	}

	project := &model.Project{
		Name:              req.Name,
		Description:       req.Description,
		AccessMode:        accessMode,
		PublicPermissions: string(publicPerms),
		GitRepoURL:        req.GitRepoURL,
		GitBranch:         gitBranch,
		CreatedBy:         userID,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, nil, err
	}

	// 创建默认环境
	defaultEnvs := []string{"dev", "test", "staging", "prod"}
	for i, env := range defaultEnvs {
		s.projectRepo.CreateEnvironment(ctx, &model.ProjectEnvironment{
			ProjectID: project.ID,
			Name:      env,
			SortOrder: i,
		})
	}

	// 创建默认 API Key
	key, err := s.createDefaultKey(ctx, project.ID)
	if err != nil {
		return project, nil, err
	}

	return project, key, nil
}

// createDefaultKey 创建默认密钥
func (s *ProjectService) createDefaultKey(ctx context.Context, projectID int64) (*model.ProjectKey, error) {
	accessKey := "ak_" + uuid.New().String()[:24]
	secretKey := "sk_" + uuid.New().String()

	secretHash, err := bcrypt.GenerateFromPassword([]byte(secretKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	perms, _ := json.Marshal(model.DefaultPermissions())

	key := &model.ProjectKey{
		ProjectID:     projectID,
		Name:          "默认密钥",
		AccessKey:     accessKey,
		SecretKeyHash: string(secretHash),
		Permissions:   string(perms),
		IsActive:      true,
	}

	if err := s.keyRepo.Create(ctx, key); err != nil {
		return nil, err
	}

	// 返回时包含明文 secret key (仅此一次)
	return &model.ProjectKey{
		ID:          key.ID,
		ProjectID:   key.ProjectID,
		Name:        key.Name,
		AccessKey:   accessKey,
		Permissions: key.Permissions,
		IsActive:    key.IsActive,
		CreatedAt:   key.CreatedAt,
	}, nil
}

// GetByID 根据 ID 获取项目
func (s *ProjectService) GetByID(ctx context.Context, id int64) (*model.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProjectNotFound
	}
	return project, nil
}

// List 获取项目列表
func (s *ProjectService) List(ctx context.Context, userID int64) ([]*model.Project, error) {
	return s.projectRepo.List(ctx, userID)
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AccessMode  string `json:"access_mode"`
	GitRepoURL  string `json:"git_repo_url"`
	GitBranch   string `json:"git_branch"`
}

// Update 更新项目
func (s *ProjectService) Update(ctx context.Context, id int64, req *UpdateProjectRequest) error {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return ErrProjectNotFound
	}

	// 检查名称是否被其他项目使用
	if req.Name != "" && req.Name != project.Name {
		exists, err := s.projectRepo.ExistsByName(ctx, req.Name)
		if err != nil {
			return err
		}
		if exists {
			return ErrProjectNameExists
		}
		project.Name = req.Name
	}

	if req.Description != "" {
		project.Description = req.Description
	}
	if req.AccessMode != "" {
		project.AccessMode = req.AccessMode
	}
	if req.GitRepoURL != "" {
		project.GitRepoURL = req.GitRepoURL
	}
	if req.GitBranch != "" {
		project.GitBranch = req.GitBranch
	}

	return s.projectRepo.Update(ctx, project)
}

// Delete 删除项目
func (s *ProjectService) Delete(ctx context.Context, id int64) error {
	_, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return ErrProjectNotFound
	}
	return s.projectRepo.Delete(ctx, id)
}

// ListEnvironments 获取项目环境列表
func (s *ProjectService) ListEnvironments(ctx context.Context, projectID int64) ([]*model.ProjectEnvironment, error) {
	return s.projectRepo.ListEnvironments(ctx, projectID)
}

// CreateEnvironment 创建环境
func (s *ProjectService) CreateEnvironment(ctx context.Context, projectID int64, name, description string) error {
	env := &model.ProjectEnvironment{
		ProjectID:   projectID,
		Name:        name,
		Description: description,
	}
	return s.projectRepo.CreateEnvironment(ctx, env)
}
