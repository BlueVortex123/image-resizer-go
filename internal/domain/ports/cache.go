package ports

import "context"

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, string, bool, error)
	Set(ctx context.Context, key string, value []byte, contentType string) error
}
