package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"confighub/internal/model"
	"confighub/internal/repository"
)

var (
	ErrVersionNotFound = errors.New("版本不存在")
)

// VersionService 版本服务
type VersionService struct {
	versionRepo *repository.VersionRepository
	configRepo  *repository.ConfigRepository
}

// NewVersionService 创建版本服务
func NewVersionService(versionRepo *repository.VersionRepository, configRepo *repository.ConfigRepository) *VersionService {
	return &VersionService{
		versionRepo: versionRepo,
		configRepo:  configRepo,
	}
}

// List 获取版本列表
func (s *VersionService) List(ctx context.Context, configID int64) ([]*model.ConfigVersion, error) {
	return s.versionRepo.List(ctx, configID)
}

// GetByVersion 获取指定版本
func (s *VersionService) GetByVersion(ctx context.Context, configID int64, version int) (*model.ConfigVersion, error) {
	v, err := s.versionRepo.GetByConfigAndVersion(ctx, configID, version)
	if err != nil {
		return nil, ErrVersionNotFound
	}
	return v, nil
}

// DiffResult 对比结果
type DiffResult struct {
	FromVersion int        `json:"from_version"`
	ToVersion   int        `json:"to_version"`
	Changes     []DiffLine `json:"changes"`
}

// DiffLine 对比行
type DiffLine struct {
	Type    string `json:"type"` // add, remove, unchanged
	LineNum int    `json:"line_num"`
	Content string `json:"content"`
}


// Diff 版本对比
func (s *VersionService) Diff(ctx context.Context, configID int64, fromV, toV int) (*DiffResult, error) {
	fromVersion, err := s.versionRepo.GetByConfigAndVersion(ctx, configID, fromV)
	if err != nil {
		return nil, ErrVersionNotFound
	}

	toVersion, err := s.versionRepo.GetByConfigAndVersion(ctx, configID, toV)
	if err != nil {
		return nil, ErrVersionNotFound
	}

	changes := diffContent(fromVersion.Content, toVersion.Content)

	return &DiffResult{
		FromVersion: fromV,
		ToVersion:   toV,
		Changes:     changes,
	}, nil
}

// Rollback 回滚到指定版本
func (s *VersionService) Rollback(ctx context.Context, configID int64, toVersion int, author string) (*model.ConfigVersion, error) {
	// 获取目标版本
	targetVersion, err := s.versionRepo.GetByConfigAndVersion(ctx, configID, toVersion)
	if err != nil {
		return nil, ErrVersionNotFound
	}

	// 获取当前配置
	config, err := s.configRepo.GetByID(ctx, configID)
	if err != nil {
		return nil, ErrConfigNotFound
	}

	// 创建新版本（内容为目标版本的内容）
	newVersionNum := config.CurrentVersion + 1
	commitHash := GenerateHash(targetVersion.Content)

	newVersion := &model.ConfigVersion{
		ConfigID:      configID,
		Version:       newVersionNum,
		Content:       targetVersion.Content,
		CommitHash:    commitHash,
		CommitMessage: "回滚到版本 " + string(rune(toVersion+'0')),
		Author:        author,
	}

	if err := s.versionRepo.Create(ctx, newVersion); err != nil {
		return nil, err
	}

	// 更新配置的当前版本
	config.CurrentVersion = newVersionNum
	if err := s.configRepo.Update(ctx, config); err != nil {
		return nil, err
	}

	return newVersion, nil
}

// GenerateHash 生成内容哈希
func GenerateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])[:16]
}

// diffContent 简单的行级对比
func diffContent(oldContent, newContent string) []DiffLine {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	var changes []DiffLine
	lineNum := 0

	// 简单的逐行对比
	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	for i := 0; i < maxLen; i++ {
		lineNum++
		var oldLine, newLine string
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine == newLine {
			changes = append(changes, DiffLine{
				Type:    "unchanged",
				LineNum: lineNum,
				Content: newLine,
			})
		} else {
			if oldLine != "" {
				changes = append(changes, DiffLine{
					Type:    "remove",
					LineNum: lineNum,
					Content: oldLine,
				})
			}
			if newLine != "" {
				changes = append(changes, DiffLine{
					Type:    "add",
					LineNum: lineNum,
					Content: newLine,
				})
			}
		}
	}

	return changes
}
