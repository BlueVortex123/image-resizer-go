package ports

import "context"

type Imager interface {
	// Resize function must accept a context, source URL, width and height
	// It returns the resized image bytes, contentType (e.g. "image/png") and an error if any
	Resize(ctx context.Context, srcURL string, w, h int) (data []byte, contentType string, err error)
}
