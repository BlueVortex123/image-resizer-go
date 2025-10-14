// internal/service/resizer.go
package service

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"example.com/img-resizer/internal/adapters/queue"
	"example.com/img-resizer/internal/config"
	"example.com/img-resizer/internal/domain/ports"
)

type ResizerService struct {
	cache       ports.Cache
	ttl         time.Duration
	config      config.Config
	queueClient *queue.QueueClient // Optional queue client for async processing
}

func NewResizerService(c ports.Cache, ttl time.Duration, cfg config.Config) *ResizerService {
	return &ResizerService{cache: c, ttl: ttl, config: cfg, queueClient: nil}
}

// NewResizerServiceWithQueue creates a resizer service with queue support for async processing
func NewResizerServiceWithQueue(c ports.Cache, ttl time.Duration, cfg config.Config, queueClient *queue.QueueClient) *ResizerService {
	return &ResizerService{cache: c, ttl: ttl, config: cfg, queueClient: queueClient}
}

func (s *ResizerService) keyFor(url string, w, h int) string {
	hsh := sha1.Sum(fmt.Appendf(nil, "%s|%d|%d", url, w, h))
	return "imgcache:" + hex.EncodeToString(hsh[:])
}

// EnqueueResizeTask adds a resize task to the queue for async processing
func (s *ResizerService) EnqueueResizeTask(url string, width, height int) error {
	if s.queueClient == nil {
		return fmt.Errorf("queue client not configured")
	}
	return s.queueClient.EnqueueResizeImageTask(url, width, height)
}

// SupportsAsync returns true if the service supports async processing
func (s *ResizerService) SupportsAsync() bool {
	return s.queueClient != nil
}
