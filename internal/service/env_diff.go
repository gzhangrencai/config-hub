package service

import (
	"context"
	"encoding/json"
	"sort"

	"confighub/internal/repository"
)

// EnvDiffService 环境对比服务
type EnvDiffService struct {
	configRepo  *repository.ConfigRepository
	versionRepo *repository.VersionRepository
}

// NewEnvDiffService 创建环境对比服务
func NewEnvDiffService(configRepo *repository.ConfigRepository, versionRepo *repository.VersionRepository) *EnvDiffService {
	return &EnvDiffService{
		configRepo:  configRepo,
		versionRepo: versionRepo,
	}
}

// EnvComparison 环境对比结果
type EnvComparison struct {
	SourceEnv    string           `json:"source_env"`
	TargetEnv    string           `json:"target_env"`
	Differences  []EnvDifference  `json:"differences"`
	OnlyInSource []string         `json:"only_in_source"`
	OnlyInTarget []string         `json:"only_in_target"`
	Summary      ComparisonSummary `json:"summary"`
}

// EnvDifference 环境差异
type EnvDifference struct {
	Path        string      `json:"path"`
	SourceValue interface{} `json:"source_value"`
	TargetValue interface{} `json:"target_value"`
	Type        string      `json:"type"` // modified, added, removed
}

// ComparisonSummary 对比摘要
type ComparisonSummary struct {
	TotalKeys    int `json:"total_keys"`
	ModifiedKeys int `json:"modified_keys"`
	AddedKeys    int `json:"added_keys"`
	RemovedKeys  int `json:"removed_keys"`
}

// Compare 对比两个环境的配置
func (s *EnvDiffService) Compare(ctx context.Context, configID int64, sourceEnv, targetEnv string) (*EnvComparison, error) {
	config, err := s.configRepo.GetByID(ctx, configID)
	if err != nil {
		return nil, ErrConfigNotFound
	}

	// 获取源环境配置
	sourceConfig, err := s.configRepo.GetByNameAndEnv(ctx, config.ProjectID, config.Name, config.Namespace, sourceEnv)
	if err != nil {
		return nil, ErrConfigNotFound
	}
	sourceVersion, err := s.versionRepo.GetLatest(ctx, sourceConfig.ID)
	if err != nil {
		return nil, err
	}

	// 获取目标环境配置
	targetConfig, err := s.configRepo.GetByNameAndEnv(ctx, config.ProjectID, config.Name, config.Namespace, targetEnv)
	if err != nil {
		return nil, ErrConfigNotFound
	}
	targetVersion, err := s.versionRepo.GetLatest(ctx, targetConfig.ID)
	if err != nil {
		return nil, err
	}

	return s.compareContent(sourceEnv, targetEnv, sourceVersion.Content, targetVersion.Content)
}

// compareContent 对比两个配置内容
func (s *EnvDiffService) compareContent(sourceEnv, targetEnv, sourceContent, targetContent string) (*EnvComparison, error) {
	var sourceData, targetData map[string]interface{}

	if err := json.Unmarshal([]byte(sourceContent), &sourceData); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(targetContent), &targetData); err != nil {
		return nil, err
	}

	comparison := &EnvComparison{
		SourceEnv:    sourceEnv,
		TargetEnv:    targetEnv,
		Differences:  []EnvDifference{},
		OnlyInSource: []string{},
		OnlyInTarget: []string{},
	}

	s.compareMap("", sourceData, targetData, comparison)

	// 计算摘要
	comparison.Summary = ComparisonSummary{
		TotalKeys:    len(comparison.Differences) + len(comparison.OnlyInSource) + len(comparison.OnlyInTarget),
		ModifiedKeys: 0,
		AddedKeys:    len(comparison.OnlyInTarget),
		RemovedKeys:  len(comparison.OnlyInSource),
	}
	for _, diff := range comparison.Differences {
		if diff.Type == "modified" {
			comparison.Summary.ModifiedKeys++
		}
	}

	return comparison, nil
}

// compareMap 递归对比 map
func (s *EnvDiffService) compareMap(prefix string, source, target map[string]interface{}, result *EnvComparison) {
	allKeys := make(map[string]bool)
	for k := range source {
		allKeys[k] = true
	}
	for k := range target {
		allKeys[k] = true
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		sourceVal, sourceExists := source[key]
		targetVal, targetExists := target[key]

		if !sourceExists {
			result.OnlyInTarget = append(result.OnlyInTarget, path)
			continue
		}
		if !targetExists {
			result.OnlyInSource = append(result.OnlyInSource, path)
			continue
		}

		sourceMap, sourceIsMap := sourceVal.(map[string]interface{})
		targetMap, targetIsMap := targetVal.(map[string]interface{})

		if sourceIsMap && targetIsMap {
			s.compareMap(path, sourceMap, targetMap, result)
		} else if !s.valuesEqual(sourceVal, targetVal) {
			result.Differences = append(result.Differences, EnvDifference{
				Path:        path,
				SourceValue: sourceVal,
				TargetValue: targetVal,
				Type:        "modified",
			})
		}
	}
}

// valuesEqual 比较两个值是否相等
func (s *EnvDiffService) valuesEqual(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

// Sync 同步配置到目标环境
func (s *EnvDiffService) Sync(ctx context.Context, configID int64, sourceEnv, targetEnv string, keys []string) error {
	config, err := s.configRepo.GetByID(ctx, configID)
	if err != nil {
		return ErrConfigNotFound
	}

	// 获取源环境配置
	sourceConfig, err := s.configRepo.GetByNameAndEnv(ctx, config.ProjectID, config.Name, config.Namespace, sourceEnv)
	if err != nil {
		return ErrConfigNotFound
	}
	sourceVersion, err := s.versionRepo.GetLatest(ctx, sourceConfig.ID)
	if err != nil {
		return err
	}

	// 获取目标环境配置
	targetConfig, err := s.configRepo.GetByNameAndEnv(ctx, config.ProjectID, config.Name, config.Namespace, targetEnv)
	if err != nil {
		return ErrConfigNotFound
	}
	targetVersion, err := s.versionRepo.GetLatest(ctx, targetConfig.ID)
	if err != nil {
		return err
	}

	var sourceData, targetData map[string]interface{}
	json.Unmarshal([]byte(sourceVersion.Content), &sourceData)
	json.Unmarshal([]byte(targetVersion.Content), &targetData)

	// 同步指定的 keys
	if len(keys) == 0 {
		targetData = sourceData
	} else {
		for _, key := range keys {
			if val, exists := sourceData[key]; exists {
				targetData[key] = val
			}
		}
	}

	newContent, _ := json.MarshalIndent(targetData, "", "  ")
	targetVersion.Content = string(newContent)

	return s.versionRepo.Update(ctx, targetVersion)
}

