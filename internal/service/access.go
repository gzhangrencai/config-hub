package service

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	"confighub/internal/model"
	"confighub/internal/repository"
)

var (
	ErrAccessDenied     = errors.New("访问被拒绝")
	ErrIPNotAllowed     = errors.New("IP 地址不在白名单中")
	ErrKeyExpired       = errors.New("密钥已过期")
	ErrKeyInactive      = errors.New("密钥已停用")
	ErrPermissionDenied = errors.New("权限不足")
)

// AccessService 访问控制服务
type AccessService struct {
	keyRepo     *repository.KeyRepository
	projectRepo *repository.ProjectRepository
}

// NewAccessService 创建访问控制服务
func NewAccessService(keyRepo *repository.KeyRepository, projectRepo *repository.ProjectRepository) *AccessService {
	return &AccessService{
		keyRepo:     keyRepo,
		projectRepo: projectRepo,
	}
}

// ValidateAccess 验证访问权限
func (s *AccessService) ValidateAccess(ctx context.Context, accessKey, clientIP string, requiredPermission string) (*model.ProjectKey, error) {
	// 获取密钥
	key, err := s.keyRepo.GetByAccessKey(ctx, accessKey)
	if err != nil {
		return nil, ErrAccessDenied
	}

	// 检查密钥是否激活
	if !key.IsActive {
		return nil, ErrKeyInactive
	}

	// 检查过期时间
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return nil, ErrKeyExpired
	}

	// 检查 IP 白名单
	if err := s.checkIPWhitelist(key, clientIP); err != nil {
		return nil, err
	}

	// 检查权限
	if requiredPermission != "" {
		if err := s.checkPermission(key, requiredPermission); err != nil {
			return nil, err
		}
	}

	return key, nil
}

// checkIPWhitelist 检查 IP 白名单
func (s *AccessService) checkIPWhitelist(key *model.ProjectKey, clientIP string) error {
	if key.IPWhitelist == "" || key.IPWhitelist == "[]" {
		return nil // 没有白名单限制
	}

	var whitelist []string
	if err := json.Unmarshal([]byte(key.IPWhitelist), &whitelist); err != nil {
		return nil // 解析失败，不限制
	}

	if len(whitelist) == 0 {
		return nil
	}

	for _, allowed := range whitelist {
		if s.matchIP(clientIP, allowed) {
			return nil
		}
	}

	return ErrIPNotAllowed
}

// matchIP 匹配 IP 地址
func (s *AccessService) matchIP(clientIP, pattern string) bool {
	// 精确匹配
	if clientIP == pattern {
		return true
	}

	// CIDR 匹配
	if strings.Contains(pattern, "/") {
		_, ipNet, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		ip := net.ParseIP(clientIP)
		if ip == nil {
			return false
		}
		return ipNet.Contains(ip)
	}

	// 通配符匹配 (如 192.168.1.*)
	if strings.Contains(pattern, "*") {
		patternParts := strings.Split(pattern, ".")
		ipParts := strings.Split(clientIP, ".")
		if len(patternParts) != len(ipParts) {
			return false
		}
		for i, part := range patternParts {
			if part != "*" && part != ipParts[i] {
				return false
			}
		}
		return true
	}

	return false
}

// checkPermission 检查权限
func (s *AccessService) checkPermission(key *model.ProjectKey, permission string) error {
	var perms model.Permissions
	if key.Permissions != "" {
		if err := json.Unmarshal([]byte(key.Permissions), &perms); err != nil {
			return ErrPermissionDenied
		}
	}

	allowed := false
	switch permission {
	case "read":
		allowed = perms.Read
	case "write":
		allowed = perms.Write
	case "delete":
		allowed = perms.Delete
	case "release":
		allowed = perms.Release
	case "admin":
		allowed = perms.Admin
	case "decrypt":
		allowed = perms.Decrypt
	}

	if !allowed {
		return ErrPermissionDenied
	}

	return nil
}

// GetPermissions 获取密钥权限
func (s *AccessService) GetPermissions(key *model.ProjectKey) model.Permissions {
	var perms model.Permissions
	if key.Permissions != "" {
		json.Unmarshal([]byte(key.Permissions), &perms)
	}
	return perms
}

// HasPermission 检查是否有指定权限
func (s *AccessService) HasPermission(key *model.ProjectKey, permission string) bool {
	perms := s.GetPermissions(key)
	switch permission {
	case "read":
		return perms.Read
	case "write":
		return perms.Write
	case "delete":
		return perms.Delete
	case "release":
		return perms.Release
	case "admin":
		return perms.Admin
	case "decrypt":
		return perms.Decrypt
	default:
		return false
	}
}

// ValidateProjectAccess 验证项目访问权限
func (s *AccessService) ValidateProjectAccess(ctx context.Context, key *model.ProjectKey, projectID int64) error {
	if key.ProjectID != projectID {
		return ErrAccessDenied
	}
	return nil
}

// IsKeyValid 检查密钥是否有效
func (s *AccessService) IsKeyValid(key *model.ProjectKey) bool {
	if !key.IsActive {
		return false
	}
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}
