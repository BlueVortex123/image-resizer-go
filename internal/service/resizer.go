// internal/service/resizer.go
package service

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"example.com/img-resizer/internal/config"
	"example.com/img-resizer/internal/domain/ports"
)

type ResizerService struct {
	cache  ports.Cache
	ttl    time.Duration
	config config.Config
}

func NewResizerService(c ports.Cache, ttl time.Duration, cfg config.Config) *ResizerService {
	return &ResizerService{cache: c, ttl: ttl, config: cfg}
}

func (s *ResizerService) keyFor(url string, w, h int) string {
	hsh := sha1.Sum(fmt.Appendf(nil, "%s|%d|%d", url, w, h))
	return "imgcache:" + hex.EncodeToString(hsh[:])
}
