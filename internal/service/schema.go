package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"confighub/internal/repository"
)

var (
	ErrInvalidSchema  = errors.New("无效的 Schema 定义")
	ErrSchemaNotFound = errors.New("Schema 不存在")
)

// SchemaService Schema 服务
type SchemaService struct {
	configRepo  *repository.ConfigRepository
	versionRepo *repository.VersionRepository
}

// NewSchemaService 创建 Schema 服务
func NewSchemaService(configRepo *repository.ConfigRepository, versionRepo *repository.VersionRepository) *SchemaService {
	return &SchemaService{
		configRepo:  configRepo,
		versionRepo: versionRepo,
	}
}

// Get 获取配置的 Schema
func (s *SchemaService) Get(ctx context.Context, configID int64) (string, error) {
	config, err := s.configRepo.GetByID(ctx, configID)
	if err != nil {
		return "", ErrConfigNotFound
	}

	if config.SchemaJSON == "" {
		return "", ErrSchemaNotFound
	}

	return config.SchemaJSON, nil
}

// Update 更新配置的 Schema
func (s *SchemaService) Update(ctx context.Context, configID int64, schema string) error {
	config, err := s.configRepo.GetByID(ctx, configID)
	if err != nil {
		return ErrConfigNotFound
	}

	// 验证 Schema 是否为有效 JSON Schema
	if err := s.validateSchema(schema); err != nil {
		return err
	}

	config.SchemaJSON = schema
	return s.configRepo.Update(ctx, config)
}

// Generate 从配置内容自动生成 Schema
func (s *SchemaService) Generate(ctx context.Context, configID int64) (string, error) {
	config, err := s.configRepo.GetByID(ctx, configID)
	if err != nil {
		return "", ErrConfigNotFound
	}

	// 获取最新版本内容
	version, err := s.versionRepo.GetLatest(ctx, configID)
	if err != nil {
		return "", errors.New("无法获取配置内容")
	}

	// 解析 JSON 内容
	var data interface{}
	if err := json.Unmarshal([]byte(version.Content), &data); err != nil {
		return "", errors.New("配置内容不是有效的 JSON")
	}

	// 生成 Schema
	schema := s.generateSchemaFromValue(data)
	schema["$schema"] = "http://json-schema.org/draft-07/schema#"

	// 保存生成的 Schema
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}

	config.SchemaJSON = string(schemaJSON)
	if err := s.configRepo.Update(ctx, config); err != nil {
		return "", err
	}

	return string(schemaJSON), nil
}

// Validate 验证内容是否符合 Schema
func (s *SchemaService) Validate(ctx context.Context, schema, content string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	// 验证 content 是否为有效 JSON
	var contentData interface{}
	if err := json.Unmarshal([]byte(content), &contentData); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "",
			Message: "内容不是有效的 JSON",
		})
		return result, nil
	}

	// 解析 Schema
	var schemaData map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &schemaData); err != nil {
		return nil, ErrInvalidSchema
	}

	// 执行验证
	errors := s.validateValue("", contentData, schemaData)
	if len(errors) > 0 {
		result.Valid = false
		result.Errors = errors
	}

	return result, nil
}

// ValidateConfig 验证配置内容是否符合其 Schema
func (s *SchemaService) ValidateConfig(ctx context.Context, configID int64, content string) (*ValidationResult, error) {
	schema, err := s.Get(ctx, configID)
	if err != nil {
		if err == ErrSchemaNotFound {
			// 没有 Schema，跳过验证
			return &ValidationResult{Valid: true}, nil
		}
		return nil, err
	}

	return s.Validate(ctx, schema, content)
}

// validateSchema 验证 Schema 是否有效
func (s *SchemaService) validateSchema(schema string) error {
	var schemaData map[string]interface{}
	if err := json.Unmarshal([]byte(schema), &schemaData); err != nil {
		return ErrInvalidSchema
	}

	// 检查必要字段
	if _, ok := schemaData["type"]; !ok {
		return errors.New("Schema 缺少 type 字段")
	}

	return nil
}

// generateSchemaFromValue 从值生成 Schema
func (s *SchemaService) generateSchemaFromValue(value interface{}) map[string]interface{} {
	schema := make(map[string]interface{})

	if value == nil {
		schema["type"] = "null"
		return schema
	}

	switch v := value.(type) {
	case bool:
		schema["type"] = "boolean"

	case float64:
		// JSON 数字默认解析为 float64
		if v == float64(int64(v)) {
			schema["type"] = "integer"
		} else {
			schema["type"] = "number"
		}

	case string:
		schema["type"] = "string"

	case []interface{}:
		schema["type"] = "array"
		if len(v) > 0 {
			// 从第一个元素推断 items 类型
			schema["items"] = s.generateSchemaFromValue(v[0])
		} else {
			schema["items"] = map[string]interface{}{}
		}

	case map[string]interface{}:
		schema["type"] = "object"
		properties := make(map[string]interface{})
		required := []string{}

		// 按键排序以保证一致性
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			properties[k] = s.generateSchemaFromValue(v[k])
			required = append(required, k)
		}

		schema["properties"] = properties
		if len(required) > 0 {
			schema["required"] = required
		}

	default:
		// 未知类型
		schema["type"] = "object"
	}

	return schema
}

// validateValue 验证值是否符合 Schema
func (s *SchemaService) validateValue(path string, value interface{}, schema map[string]interface{}) []ValidationError {
	var errors []ValidationError

	schemaType, _ := schema["type"].(string)

	// 类型检查
	if schemaType != "" {
		actualType := s.getJSONType(value)
		if !s.typeMatches(actualType, schemaType) {
			errors = append(errors, ValidationError{
				Field:   path,
				Message: fmt.Sprintf("类型不匹配: 期望 %s, 实际 %s", schemaType, actualType),
			})
			return errors
		}
	}

	switch schemaType {
	case "object":
		obj, ok := value.(map[string]interface{})
		if !ok {
			break
		}

		// 检查 required 字段
		if required, ok := schema["required"].([]interface{}); ok {
			for _, r := range required {
				fieldName := r.(string)
				if _, exists := obj[fieldName]; !exists {
					fieldPath := fieldName
					if path != "" {
						fieldPath = path + "." + fieldName
					}
					errors = append(errors, ValidationError{
						Field:   fieldPath,
						Message: "缺少必填字段",
					})
				}
			}
		}

		// 验证 properties
		if properties, ok := schema["properties"].(map[string]interface{}); ok {
			for key, propSchema := range properties {
				if propValue, exists := obj[key]; exists {
					fieldPath := key
					if path != "" {
						fieldPath = path + "." + key
					}
					if ps, ok := propSchema.(map[string]interface{}); ok {
						errors = append(errors, s.validateValue(fieldPath, propValue, ps)...)
					}
				}
			}
		}

	case "array":
		arr, ok := value.([]interface{})
		if !ok {
			break
		}

		// 检查 minItems
		if minItems, ok := schema["minItems"].(float64); ok {
			if float64(len(arr)) < minItems {
				errors = append(errors, ValidationError{
					Field:   path,
					Message: fmt.Sprintf("数组长度不能小于 %d", int(minItems)),
				})
			}
		}

		// 检查 maxItems
		if maxItems, ok := schema["maxItems"].(float64); ok {
			if float64(len(arr)) > maxItems {
				errors = append(errors, ValidationError{
					Field:   path,
					Message: fmt.Sprintf("数组长度不能大于 %d", int(maxItems)),
				})
			}
		}

		// 验证 items
		if itemSchema, ok := schema["items"].(map[string]interface{}); ok {
			for i, item := range arr {
				itemPath := fmt.Sprintf("%s[%d]", path, i)
				if path == "" {
					itemPath = fmt.Sprintf("[%d]", i)
				}
				errors = append(errors, s.validateValue(itemPath, item, itemSchema)...)
			}
		}

	case "string":
		str, ok := value.(string)
		if !ok {
			break
		}

		// 检查 minLength
		if minLength, ok := schema["minLength"].(float64); ok {
			if float64(len(str)) < minLength {
				errors = append(errors, ValidationError{
					Field:   path,
					Message: fmt.Sprintf("字符串长度不能小于 %d", int(minLength)),
				})
			}
		}

		// 检查 maxLength
		if maxLength, ok := schema["maxLength"].(float64); ok {
			if float64(len(str)) > maxLength {
				errors = append(errors, ValidationError{
					Field:   path,
					Message: fmt.Sprintf("字符串长度不能大于 %d", int(maxLength)),
				})
			}
		}

		// 检查 enum
		if enum, ok := schema["enum"].([]interface{}); ok {
			found := false
			for _, e := range enum {
				if e == str {
					found = true
					break
				}
			}
			if !found {
				errors = append(errors, ValidationError{
					Field:   path,
					Message: fmt.Sprintf("值必须是以下之一: %v", enum),
				})
			}
		}

	case "number", "integer":
		num, ok := value.(float64)
		if !ok {
			break
		}

		// 检查 minimum
		if minimum, ok := schema["minimum"].(float64); ok {
			if num < minimum {
				errors = append(errors, ValidationError{
					Field:   path,
					Message: fmt.Sprintf("值不能小于 %v", minimum),
				})
			}
		}

		// 检查 maximum
		if maximum, ok := schema["maximum"].(float64); ok {
			if num > maximum {
				errors = append(errors, ValidationError{
					Field:   path,
					Message: fmt.Sprintf("值不能大于 %v", maximum),
				})
			}
		}
	}

	return errors
}

// getJSONType 获取值的 JSON 类型
func (s *SchemaService) getJSONType(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch value.(type) {
	case bool:
		return "boolean"
	case float64:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return reflect.TypeOf(value).String()
	}
}

// typeMatches 检查类型是否匹配
func (s *SchemaService) typeMatches(actual, expected string) bool {
	if actual == expected {
		return true
	}
	// integer 也是 number
	if expected == "number" && actual == "number" {
		return true
	}
	if expected == "integer" && actual == "number" {
		return true
	}
	return false
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
