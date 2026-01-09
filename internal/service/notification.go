package service

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
)

// NotificationService 通知服务
type NotificationService struct {
	rdb         *redis.Client
	subscribers map[string]chan *ConfigChange
	mu          sync.RWMutex
}

// ConfigChange 配置变更
type ConfigChange struct {
	ConfigID   int64  `json:"config_id"`
	ConfigName string `json:"config_name"`
	Namespace  string `json:"namespace"`
	Env        string `json:"environment"`
	Version    int    `json:"version"`
	ChangeType string `json:"change_type"`
}

// NewNotificationService 创建通知服务
func NewNotificationService(rdb *redis.Client) *NotificationService {
	return &NotificationService{
		rdb:         rdb,
		subscribers: make(map[string]chan *ConfigChange),
	}
}

// Subscribe 订阅配置变更
func (s *NotificationService) Subscribe(ctx context.Context, clientID string, configIDs []int64) (<-chan *ConfigChange, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *ConfigChange, 10)
	s.subscribers[clientID] = ch

	return ch, nil
}

// Unsubscribe 取消订阅
func (s *NotificationService) Unsubscribe(ctx context.Context, clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.subscribers[clientID]; ok {
		close(ch)
		delete(s.subscribers, clientID)
	}
}

// NotifyChange 通知配置变更
func (s *NotificationService) NotifyChange(ctx context.Context, change *ConfigChange) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.subscribers {
		select {
		case ch <- change:
		default:
			// 通道已满，跳过
		}
	}

	return nil
}
