package service

import (
	"context"
	"encoding/json"
	"errors"
	"hash/fnv"
	"net"
	"strings"

	"confighub/internal/model"
	"confighub/internal/repository"
)

var (
	ErrGrayReleaseNotFound = errors.New("灰度发布不存在")
	ErrGrayReleaseActive   = errors.New("已有活跃的灰度发布")
)

// GrayReleaseService 灰度发布服务
type GrayReleaseService struct {
	releaseRepo *repository.ReleaseRepository
	configRepo  *repository.ConfigRepository
	versionRepo *repository.VersionRepository
}

// NewGrayReleaseService 创建灰度发布服务
func NewGrayReleaseService(releaseRepo *repository.ReleaseRepository, configRepo *repository.ConfigRepository, versionRepo *repository.VersionRepository) *GrayReleaseService {
	return &GrayReleaseService{
		releaseRepo: releaseRepo,
		configRepo:  configRepo,
		versionRepo: versionRepo,
	}
}

// GrayReleaseRequest 灰度发布请求
type GrayReleaseRequest struct {
	ConfigID    int64    `json:"config_id"`
	Environment string   `json:"environment"`
	Version     int      `json:"version"`
	RuleType    string   `json:"rule_type"` // percentage, client_id, ip_range
	Percentage  int      `json:"percentage,omitempty"`
	ClientIDs   []string `json:"client_ids,omitempty"`
	IPRanges    []string `json:"ip_ranges,omitempty"`
}

// Create 创建灰度发布
func (s *GrayReleaseService) Create(ctx context.Context, req *GrayReleaseRequest, author string) (*model.Release, error) {
	config, err := s.configRepo.GetByID(ctx, req.ConfigID)
	if err != nil {
		return nil, ErrConfigNotFound
	}

	// 检查是否已有活跃的灰度发布
	existing, _ := s.releaseRepo.GetActiveGrayRelease(ctx, req.ConfigID, req.Environment)
	if existing != nil {
		return nil, ErrGrayReleaseActive
	}

	// 验证版本存在
	if req.Version == 0 {
		req.Version = config.CurrentVersion
	}
	_, err = s.versionRepo.GetByConfigAndVersion(ctx, req.ConfigID, req.Version)
	if err != nil {
		return nil, ErrVersionNotFound
	}

	// 构建灰度规则
	rules := model.GrayRules{
		Type:       req.RuleType,
		Percentage: req.Percentage,
		ClientIDs:  req.ClientIDs,
		IPRanges:   req.IPRanges,
	}
	rulesJSON, _ := json.Marshal(rules)

	release := &model.Release{
		ProjectID:      config.ProjectID,
		ConfigID:       req.ConfigID,
		Version:        req.Version,
		Environment:    req.Environment,
		Status:         "gray",
		ReleaseType:    "gray",
		GrayRules:      string(rulesJSON),
		GrayPercentage: req.Percentage,
		ReleasedBy:     author,
	}

	if err := s.releaseRepo.Create(ctx, release); err != nil {
		return nil, err
	}

	return release, nil
}

// ShouldUseGrayRelease 判断客户端是否应该使用灰度版本
func (s *GrayReleaseService) ShouldUseGrayRelease(ctx context.Context, configID int64, env, clientID, clientIP string) (bool, *model.Release, error) {
	grayRelease, err := s.releaseRepo.GetActiveGrayRelease(ctx, configID, env)
	if err != nil {
		return false, nil, nil // 没有灰度发布
	}

	var rules model.GrayRules
	if err := json.Unmarshal([]byte(grayRelease.GrayRules), &rules); err != nil {
		return false, nil, nil
	}

	shouldUse := false
	switch rules.Type {
	case "percentage":
		shouldUse = s.matchPercentage(clientID, rules.Percentage)
	case "client_id":
		shouldUse = s.matchClientID(clientID, rules.ClientIDs)
	case "ip_range":
		shouldUse = s.matchIPRange(clientIP, rules.IPRanges)
	}

	if shouldUse {
		return true, grayRelease, nil
	}
	return false, nil, nil
}

// matchPercentage 百分比匹配
func (s *GrayReleaseService) matchPercentage(clientID string, percentage int) bool {
	if percentage <= 0 {
		return false
	}
	if percentage >= 100 {
		return true
	}

	h := fnv.New32a()
	h.Write([]byte(clientID))
	hash := h.Sum32()
	return int(hash%100) < percentage
}

// matchClientID 客户端 ID 匹配
func (s *GrayReleaseService) matchClientID(clientID string, allowedIDs []string) bool {
	for _, id := range allowedIDs {
		if id == clientID {
			return true
		}
		// 支持通配符
		if strings.HasSuffix(id, "*") {
			prefix := strings.TrimSuffix(id, "*")
			if strings.HasPrefix(clientID, prefix) {
				return true
			}
		}
	}
	return false
}

// matchIPRange IP 范围匹配
func (s *GrayReleaseService) matchIPRange(clientIP string, ipRanges []string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	for _, rangeStr := range ipRanges {
		if strings.Contains(rangeStr, "/") {
			_, ipNet, err := net.ParseCIDR(rangeStr)
			if err == nil && ipNet.Contains(ip) {
				return true
			}
		} else if rangeStr == clientIP {
			return true
		}
	}
	return false
}

// Promote 提升灰度发布为正式发布
func (s *GrayReleaseService) Promote(ctx context.Context, releaseID int64, author string) (*model.Release, error) {
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		return nil, ErrGrayReleaseNotFound
	}

	if release.Status != "gray" {
		return nil, errors.New("只能提升灰度发布")
	}

	// 更新灰度发布状态
	release.Status = "promoted"
	if err := s.releaseRepo.Update(ctx, release); err != nil {
		return nil, err
	}

	// 创建正式发布
	fullRelease := &model.Release{
		ProjectID:   release.ProjectID,
		ConfigID:    release.ConfigID,
		Version:     release.Version,
		Environment: release.Environment,
		Status:      "released",
		ReleaseType: "full",
		ReleasedBy:  author,
	}

	if err := s.releaseRepo.Create(ctx, fullRelease); err != nil {
		return nil, err
	}

	return fullRelease, nil
}

// Cancel 取消灰度发布
func (s *GrayReleaseService) Cancel(ctx context.Context, releaseID int64) error {
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		return ErrGrayReleaseNotFound
	}

	if release.Status != "gray" {
		return errors.New("只能取消灰度发布")
	}

	release.Status = "cancelled"
	return s.releaseRepo.Update(ctx, release)
}

// UpdatePercentage 更新灰度百分比
func (s *GrayReleaseService) UpdatePercentage(ctx context.Context, releaseID int64, percentage int) error {
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		return ErrGrayReleaseNotFound
	}

	if release.Status != "gray" {
		return errors.New("只能更新灰度发布")
	}

	var rules model.GrayRules
	json.Unmarshal([]byte(release.GrayRules), &rules)
	rules.Percentage = percentage
	rulesJSON, _ := json.Marshal(rules)

	release.GrayRules = string(rulesJSON)
	release.GrayPercentage = percentage

	return s.releaseRepo.Update(ctx, release)
}

// GetGrayReleaseVersion 获取灰度发布版本内容
func (s *GrayReleaseService) GetGrayReleaseVersion(ctx context.Context, release *model.Release) (*model.ConfigVersion, error) {
	return s.versionRepo.GetByConfigAndVersion(ctx, release.ConfigID, release.Version)
}

