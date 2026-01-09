package service

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// HashService 哈希服务
type HashService struct{}

// NewHashService 创建哈希服务
func NewHashService() *HashService {
	return &HashService{}
}

// GenerateCommitHash 生成提交哈希
// 使用内容的 SHA256 哈希，取前 16 个字符
func (s *HashService) GenerateCommitHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])[:16]
}

// GenerateFullHash 生成完整的 SHA256 哈希
func (s *HashService) GenerateFullHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// GenerateTimestampedHash 生成带时间戳的哈希
// 用于确保即使内容相同，不同时间生成的哈希也不同
func (s *HashService) GenerateTimestampedHash(content string) string {
	data := content + time.Now().String()
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

// VerifyHash 验证内容哈希是否匹配
func (s *HashService) VerifyHash(content, expectedHash string) bool {
	actualHash := s.GenerateCommitHash(content)
	return actualHash == expectedHash
}

// CompareContent 比较两个内容是否相同（通过哈希）
func (s *HashService) CompareContent(content1, content2 string) bool {
	hash1 := s.GenerateFullHash(content1)
	hash2 := s.GenerateFullHash(content2)
	return hash1 == hash2
}
