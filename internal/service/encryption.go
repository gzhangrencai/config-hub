package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

var (
	ErrEncryptionFailed = errors.New("加密失败")
	ErrDecryptionFailed = errors.New("解密失败")
)

const (
	// EncryptedPrefix 加密值前缀
	EncryptedPrefix = "ENC:"
)

// EncryptionService 加密服务
type EncryptionService struct {
	key []byte
}

// NewEncryptionService 创建加密服务
func NewEncryptionService(key string) *EncryptionService {
	keyBytes := []byte(key)
	if len(keyBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded, keyBytes)
		keyBytes = padded
	} else if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	return &EncryptionService{
		key: keyBytes,
	}
}

// Encrypt 加密
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", ErrEncryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", ErrEncryptionFailed
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", ErrEncryptionFailed
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密
func (s *EncryptionService) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrDecryptionFailed
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// EncryptWithPrefix 加密并添加前缀
func (s *EncryptionService) EncryptWithPrefix(plaintext string) (string, error) {
	encrypted, err := s.Encrypt(plaintext)
	if err != nil {
		return "", err
	}
	return EncryptedPrefix + encrypted, nil
}

// DecryptWithPrefix 解密带前缀的值
func (s *EncryptionService) DecryptWithPrefix(value string) (string, error) {
	if !strings.HasPrefix(value, EncryptedPrefix) {
		return value, nil // 不是加密值，直接返回
	}
	return s.Decrypt(value[len(EncryptedPrefix):])
}

// IsEncrypted 检查值是否已加密
func (s *EncryptionService) IsEncrypted(value string) bool {
	return strings.HasPrefix(value, EncryptedPrefix)
}

// EncryptFields 加密 JSON 中的指定字段
func (s *EncryptionService) EncryptFields(content string, fields []string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", err
	}

	fieldSet := make(map[string]bool)
	for _, f := range fields {
		fieldSet[f] = true
	}

	s.encryptMapFields(data, fieldSet, "")

	result, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// DecryptFields 解密 JSON 中的所有加密字段
func (s *EncryptionService) DecryptFields(content string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", err
	}

	s.decryptMapFields(data)

	result, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// encryptMapFields 递归加密 map 中的指定字段
func (s *EncryptionService) encryptMapFields(data map[string]interface{}, fields map[string]bool, prefix string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			if fields[fullKey] || fields[key] {
				if !s.IsEncrypted(v) {
					encrypted, err := s.EncryptWithPrefix(v)
					if err == nil {
						data[key] = encrypted
					}
				}
			}
		case map[string]interface{}:
			s.encryptMapFields(v, fields, fullKey)
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					s.encryptMapFields(m, fields, fullKey)
				}
			}
		}
	}
}

// decryptMapFields 递归解密 map 中的加密字段
func (s *EncryptionService) decryptMapFields(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			if s.IsEncrypted(v) {
				decrypted, err := s.DecryptWithPrefix(v)
				if err == nil {
					data[key] = decrypted
				}
			}
		case map[string]interface{}:
			s.decryptMapFields(v)
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					s.decryptMapFields(m)
				}
			}
		}
	}
}

// MaskEncryptedFields 将加密字段显示为掩码
func (s *EncryptionService) MaskEncryptedFields(content string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return "", err
	}

	s.maskMapFields(data)

	result, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// maskMapFields 递归掩码 map 中的加密字段
func (s *EncryptionService) maskMapFields(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			if s.IsEncrypted(v) {
				data[key] = "******"
			}
		case map[string]interface{}:
			s.maskMapFields(v)
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					s.maskMapFields(m)
				}
			}
		}
	}
}
