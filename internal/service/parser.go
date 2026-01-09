package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrParseJSON = errors.New("JSON 解析失败")
	ErrParseYAML = errors.New("YAML 解析失败")
)

// Parser 配置解析器
type Parser struct{}

// NewParser 创建解析器
func NewParser() *Parser {
	return &Parser{}
}

// ParseResult 解析结果
type ParseResult struct {
	Valid       bool     `json:"valid"`
	Content     string   `json:"content"`      // 标准化后的内容
	FileType    string   `json:"file_type"`    // 检测到的文件类型
	Errors      []string `json:"errors,omitempty"`
}

// ParseJSON 解析并验证 JSON
func (p *Parser) ParseJSON(content string) (*ParseResult, error) {
	result := &ParseResult{
		FileType: "json",
		Errors:   []string{},
	}

	// 检查是否为有效 JSON
	if !json.Valid([]byte(content)) {
		result.Valid = false
		result.Errors = append(result.Errors, "无效的 JSON 格式")
		return result, ErrParseJSON
	}

	// 格式化 JSON
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("JSON 解析错误: %v", err))
		return result, ErrParseJSON
	}

	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("JSON 格式化错误: %v", err))
		return result, ErrParseJSON
	}

	result.Valid = true
	result.Content = string(formatted)
	return result, nil
}

// ParseYAML 解析 YAML 并转换为 JSON
func (p *Parser) ParseYAML(content string) (*ParseResult, error) {
	result := &ParseResult{
		FileType: "yaml",
		Errors:   []string{},
	}

	// 解析 YAML
	var data interface{}
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("YAML 解析错误: %v", err))
		return result, ErrParseYAML
	}

	// 转换 YAML 特殊类型
	data = convertYAMLToJSON(data)

	// 转换为 JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("JSON 转换错误: %v", err))
		return result, ErrParseYAML
	}

	result.Valid = true
	result.Content = string(jsonBytes)
	return result, nil
}

// DetectAndParse 自动检测文件类型并解析
func (p *Parser) DetectAndParse(content string) (*ParseResult, error) {
	content = strings.TrimSpace(content)

	// 尝试 JSON
	if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
		return p.ParseJSON(content)
	}

	// 尝试 YAML
	return p.ParseYAML(content)
}

// FormatJSON 格式化 JSON
func (p *Parser) FormatJSON(content string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", ErrParseJSON
	}

	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", ErrParseJSON
	}

	return string(formatted), nil
}

// MinifyJSON 压缩 JSON
func (p *Parser) MinifyJSON(content string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", ErrParseJSON
	}

	minified, err := json.Marshal(data)
	if err != nil {
		return "", ErrParseJSON
	}

	return string(minified), nil
}

// ValidateJSON 验证 JSON 格式
func (p *Parser) ValidateJSON(content string) bool {
	return json.Valid([]byte(content))
}

// convertYAMLToJSON 转换 YAML 特殊类型为 JSON 兼容类型
func convertYAMLToJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = convertYAMLToJSON(value)
		}
		return result
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			strKey := fmt.Sprintf("%v", key)
			result[strKey] = convertYAMLToJSON(value)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, value := range v {
			result[i] = convertYAMLToJSON(value)
		}
		return result
	default:
		return v
	}
}
