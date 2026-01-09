package service

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
)

// DiffService 差异对比服务
type DiffService struct{}

// NewDiffService 创建差异对比服务
func NewDiffService() *DiffService {
	return &DiffService{}
}

// LineDiff 行级差异
type LineDiff struct {
	Type       string `json:"type"`        // add, remove, unchanged
	LineNumber int    `json:"line_number"`
	OldLine    int    `json:"old_line,omitempty"`
	NewLine    int    `json:"new_line,omitempty"`
	Content    string `json:"content"`
}

// JSONDiff JSON 差异
type JSONDiff struct {
	Path     string      `json:"path"`
	Type     string      `json:"type"` // add, remove, modify
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
}

// DiffLines 行级对比
func (s *DiffService) DiffLines(oldContent, newContent string) []LineDiff {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	// 使用简单的 LCS 算法进行对比
	return s.simpleDiff(oldLines, newLines)
}

// simpleDiff 简单的行级对比
func (s *DiffService) simpleDiff(oldLines, newLines []string) []LineDiff {
	var diffs []LineDiff
	
	oldIdx, newIdx := 0, 0
	lineNum := 0

	for oldIdx < len(oldLines) || newIdx < len(newLines) {
		lineNum++

		if oldIdx >= len(oldLines) {
			// 新增行
			diffs = append(diffs, LineDiff{
				Type:       "add",
				LineNumber: lineNum,
				NewLine:    newIdx + 1,
				Content:    newLines[newIdx],
			})
			newIdx++
			continue
		}

		if newIdx >= len(newLines) {
			// 删除行
			diffs = append(diffs, LineDiff{
				Type:       "remove",
				LineNumber: lineNum,
				OldLine:    oldIdx + 1,
				Content:    oldLines[oldIdx],
			})
			oldIdx++
			continue
		}

		if oldLines[oldIdx] == newLines[newIdx] {
			// 相同行
			diffs = append(diffs, LineDiff{
				Type:       "unchanged",
				LineNumber: lineNum,
				OldLine:    oldIdx + 1,
				NewLine:    newIdx + 1,
				Content:    oldLines[oldIdx],
			})
			oldIdx++
			newIdx++
		} else {
			// 不同行 - 先删除旧行，再添加新行
			diffs = append(diffs, LineDiff{
				Type:       "remove",
				LineNumber: lineNum,
				OldLine:    oldIdx + 1,
				Content:    oldLines[oldIdx],
			})
			diffs = append(diffs, LineDiff{
				Type:       "add",
				LineNumber: lineNum,
				NewLine:    newIdx + 1,
				Content:    newLines[newIdx],
			})
			oldIdx++
			newIdx++
		}
	}

	return diffs
}


// DiffJSON JSON 结构对比
func (s *DiffService) DiffJSON(oldJSON, newJSON string) ([]JSONDiff, error) {
	var oldData, newData interface{}

	if err := json.Unmarshal([]byte(oldJSON), &oldData); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(newJSON), &newData); err != nil {
		return nil, err
	}

	var diffs []JSONDiff
	s.compareJSON("", oldData, newData, &diffs)
	return diffs, nil
}

// compareJSON 递归比较 JSON
func (s *DiffService) compareJSON(path string, oldVal, newVal interface{}, diffs *[]JSONDiff) {
	if reflect.DeepEqual(oldVal, newVal) {
		return
	}

	oldType := reflect.TypeOf(oldVal)
	newType := reflect.TypeOf(newVal)

	// 类型不同
	if oldType != newType {
		*diffs = append(*diffs, JSONDiff{
			Path:     path,
			Type:     "modify",
			OldValue: oldVal,
			NewValue: newVal,
		})
		return
	}

	switch old := oldVal.(type) {
	case map[string]interface{}:
		newMap := newVal.(map[string]interface{})
		
		// 收集所有键
		allKeys := make(map[string]bool)
		for k := range old {
			allKeys[k] = true
		}
		for k := range newMap {
			allKeys[k] = true
		}

		// 排序键
		keys := make([]string, 0, len(allKeys))
		for k := range allKeys {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			childPath := k
			if path != "" {
				childPath = path + "." + k
			}

			oldChild, oldExists := old[k]
			newChild, newExists := newMap[k]

			if !oldExists {
				*diffs = append(*diffs, JSONDiff{
					Path:     childPath,
					Type:     "add",
					NewValue: newChild,
				})
			} else if !newExists {
				*diffs = append(*diffs, JSONDiff{
					Path:     childPath,
					Type:     "remove",
					OldValue: oldChild,
				})
			} else {
				s.compareJSON(childPath, oldChild, newChild, diffs)
			}
		}

	case []interface{}:
		newArr := newVal.([]interface{})
		maxLen := len(old)
		if len(newArr) > maxLen {
			maxLen = len(newArr)
		}

		for i := 0; i < maxLen; i++ {
			childPath := path + "[" + string(rune('0'+i)) + "]"
			if i >= len(old) {
				*diffs = append(*diffs, JSONDiff{
					Path:     childPath,
					Type:     "add",
					NewValue: newArr[i],
				})
			} else if i >= len(newArr) {
				*diffs = append(*diffs, JSONDiff{
					Path:     childPath,
					Type:     "remove",
					OldValue: old[i],
				})
			} else {
				s.compareJSON(childPath, old[i], newArr[i], diffs)
			}
		}

	default:
		*diffs = append(*diffs, JSONDiff{
			Path:     path,
			Type:     "modify",
			OldValue: oldVal,
			NewValue: newVal,
		})
	}
}

// GetDiffSummary 获取差异摘要
func (s *DiffService) GetDiffSummary(diffs []LineDiff) map[string]int {
	summary := map[string]int{
		"added":     0,
		"removed":   0,
		"unchanged": 0,
	}

	for _, d := range diffs {
		switch d.Type {
		case "add":
			summary["added"]++
		case "remove":
			summary["removed"]++
		case "unchanged":
			summary["unchanged"]++
		}
	}

	return summary
}
