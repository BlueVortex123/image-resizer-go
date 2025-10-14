package ports

import (
	"context"

	"example.com/img-resizer/internal/domain/ports/requests"
)

type ResizerService interface {
	// GetResized must return resized image bytes and contentType (e.g. "image/png")
	GetResized(ctx context.Context, req requests.RequestGetResize) requests.ResponseGetResize
	
	// EnqueueResizeTask adds a resize task to the queue for async processing
	EnqueueResizeTask(url string, width, height int) error
	
	// SupportsAsync returns true if the service supports async processing
	SupportsAsync() bool
}
