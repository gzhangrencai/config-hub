package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"confighub/internal/model"
	"confighub/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrKeyNotFound = errors.New("密钥不存在")
)

// KeyService 密钥服务
type KeyService struct {
	keyRepo *repository.KeyRepository
}

// NewKeyService 创建密钥服务
func NewKeyService(keyRepo *repository.KeyRepository) *KeyService {
	return &KeyService{
		keyRepo: keyRepo,
	}
}

// CreateKeyRequest 创建密钥请求
type CreateKeyRequest struct {
	Name        string            `json:"name" binding:"required"`
	Permissions map[string]bool   `json:"permissions"`
	IPWhitelist []string          `json:"ip_whitelist"`
	ExpiresAt   *time.Time        `json:"expires_at"`
}

// Create 创建密钥
func (s *KeyService) Create(ctx context.Context, projectID int64, req *CreateKeyRequest) (*model.ProjectKey, string, error) {
	accessKey := "ak_" + uuid.New().String()[:24]
	secretKey := "sk_" + uuid.New().String()

	secretHash, err := bcrypt.GenerateFromPassword([]byte(secretKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	// 默认权限
	permissions := req.Permissions
	if permissions == nil {
		permissions = map[string]bool{
			"read":    true,
			"write":   false,
			"delete":  false,
			"release": false,
			"admin":   false,
		}
	}
	permsJSON, _ := json.Marshal(permissions)

	// IP 白名单
	var ipWhitelistJSON string
	if len(req.IPWhitelist) > 0 {
		ipJSON, _ := json.Marshal(req.IPWhitelist)
		ipWhitelistJSON = string(ipJSON)
	}

	key := &model.ProjectKey{
		ProjectID:     projectID,
		Name:          req.Name,
		AccessKey:     accessKey,
		SecretKeyHash: string(secretHash),
		Permissions:   string(permsJSON),
		IPWhitelist:   ipWhitelistJSON,
		ExpiresAt:     req.ExpiresAt,
		IsActive:      true,
	}

	if err := s.keyRepo.Create(ctx, key); err != nil {
		return nil, "", err
	}

	return key, secretKey, nil
}

// List 获取密钥列表
func (s *KeyService) List(ctx context.Context, projectID int64) ([]*model.ProjectKey, error) {
	return s.keyRepo.List(ctx, projectID)
}


// UpdateKeyRequest 更新密钥请求
type UpdateKeyRequest struct {
	Name        string          `json:"name"`
	Permissions map[string]bool `json:"permissions"`
	IPWhitelist []string        `json:"ip_whitelist"`
	ExpiresAt   *time.Time      `json:"expires_at"`
	IsActive    *bool           `json:"is_active"`
}

// Update 更新密钥
func (s *KeyService) Update(ctx context.Context, id int64, req *UpdateKeyRequest) error {
	key, err := s.keyRepo.GetByID(ctx, id)
	if err != nil {
		return ErrKeyNotFound
	}

	if req.Name != "" {
		key.Name = req.Name
	}
	if req.Permissions != nil {
		permsJSON, _ := json.Marshal(req.Permissions)
		key.Permissions = string(permsJSON)
	}
	if req.IPWhitelist != nil {
		ipJSON, _ := json.Marshal(req.IPWhitelist)
		key.IPWhitelist = string(ipJSON)
	}
	if req.ExpiresAt != nil {
		key.ExpiresAt = req.ExpiresAt
	}
	if req.IsActive != nil {
		key.IsActive = *req.IsActive
	}

	return s.keyRepo.Update(ctx, key)
}

// Delete 删除密钥
func (s *KeyService) Delete(ctx context.Context, id int64) error {
	_, err := s.keyRepo.GetByID(ctx, id)
	if err != nil {
		return ErrKeyNotFound
	}
	return s.keyRepo.Delete(ctx, id)
}

// Regenerate 重新生成密钥
func (s *KeyService) Regenerate(ctx context.Context, id int64) (*model.ProjectKey, string, error) {
	key, err := s.keyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, "", ErrKeyNotFound
	}

	// 生成新的密钥对
	newAccessKey := "ak_" + uuid.New().String()[:24]
	newSecretKey := "sk_" + uuid.New().String()

	secretHash, err := bcrypt.GenerateFromPassword([]byte(newSecretKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	key.AccessKey = newAccessKey
	key.SecretKeyHash = string(secretHash)

	if err := s.keyRepo.Update(ctx, key); err != nil {
		return nil, "", err
	}

	return key, newSecretKey, nil
}

// GetByAccessKey 根据 Access Key 获取密钥
func (s *KeyService) GetByAccessKey(ctx context.Context, accessKey string) (*model.ProjectKey, error) {
	key, err := s.keyRepo.GetByAccessKey(ctx, accessKey)
	if err != nil {
		return nil, ErrKeyNotFound
	}
	return key, nil
}

// ValidateSecretKey 验证 Secret Key
func (s *KeyService) ValidateSecretKey(key *model.ProjectKey, secretKey string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(key.SecretKeyHash), []byte(secretKey))
	return err == nil
}
