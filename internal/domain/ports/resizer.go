package ports

import (
	"context"

	"example.com/img-resizer/internal/domain/ports/requests"
)

type ResizerService interface {
	// GetResized must return resized image bytes and contentType (e.g. "image/png")
	GetResized(ctx context.Context, req requests.RequestGetResize) requests.ResponseGetResize
}
