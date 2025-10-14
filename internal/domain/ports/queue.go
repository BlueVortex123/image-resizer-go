package ports

// QueueService defines the interface for queue operations
type QueueService interface {
	// EnqueueResizeTask adds a resize task to the queue for async processing
	EnqueueResizeTask(url string, width, height int) error
	
	// Close closes the queue client connection
	Close() error
}